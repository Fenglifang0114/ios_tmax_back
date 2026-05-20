//go:build !windows
package svc

import (
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"tmaxsrv/comm"
	"tmaxsrv/log"
	"tmaxsrv/picker"
	"tmaxsrv/util"
)

type BtInfoList struct {
	Mac  string `json:"mac"`
	Name string `json:"name"`
	RSSI int    `json:"rssi"`
}

type BluetoothManager struct {
	adapter interface{}
}

type TBluetooth struct {
	recvCh           chan comm.Packet
	sendCh           chan []byte
	notificationChan chan []byte
	quitChan         chan struct{}
	wg               sync.WaitGroup
	isAlive          atomic.Bool
	toQuit           bool
	isDefault        bool
	queue            *util.CircularBuffer
	pickerFn         picker.PickerFunc

	VirtualWriteHandler func([]byte)
}

func (t *TBluetooth) Close() {
	t.toQuit = true
	if t.isAlive.Swap(false) {
		close(t.quitChan)
		t.wg.Wait()
	}
}

func GetBluetoothManager() *BluetoothManager { return &BluetoothManager{} }
func (m *BluetoothManager) EnableAdapter() error { return nil }
func (m *BluetoothManager) GetScaleList() []BtInfoList { return nil }
func (m *BluetoothManager) StopScan() {}
func (m *BluetoothManager) DisconnectAll() {}
func (m *BluetoothManager) Disconnect(address string) {}
func (m *BluetoothManager) IsConnected(address string) bool { return false }

func NewBluetoothConnection(address string, pickerFn picker.PickerFunc, isDefault bool, adapter interface{}) (*TBluetooth, error) {
	bt := &TBluetooth{
		recvCh:           make(chan comm.Packet, 100),
		sendCh:           make(chan []byte, 100),
		notificationChan: make(chan []byte, 100),
		quitChan:         make(chan struct{}),
		isDefault:        isDefault,
		queue:            util.NewCircularBuffer(2048),
		pickerFn:         pickerFn,
	}
	bt.isAlive.Store(true)

	go bt.readLoop()
	go bt.writeLoop()
	return bt, nil
}

func (bt *TBluetooth) writeLoop() {
	bt.wg.Add(1)
	defer bt.wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Log.Errorf("BT writeLoop panic recovered: %v", err)
		}
	}()

	for {
		select {
		case data := <-bt.sendCh:
			if bt.VirtualWriteHandler != nil {
				bt.VirtualWriteHandler(data)
			}
		case <-bt.quitChan:
			return
		}
	}
}

func (t *TBluetooth) Write(data []byte) error {
	select {
	case t.sendCh <- data:
		return nil
	default:
		return fmt.Errorf("sendCh full")
	}
}

func (bt *TBluetooth) readLoop() {
	bt.wg.Add(1)
	defer bt.wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Log.Errorf("BT readLoop panic recovered: %v", err)
		}
	}()

	for {
		select {
		case data := <-bt.notificationChan:
			if data != nil && len(data) > 0 {
				bt.processReceivedData(data)
			}
			bt.processQueue()
		case <-bt.quitChan:
			return
		case <-time.After(10 * time.Millisecond):
			bt.processQueue()
		}
	}
}

func (bt *TBluetooth) processReceivedData(data []byte) {
	if bt.queue.IsFull() {
		bt.queue.DequeueN(bt.queue.Capacity)
	}
	bt.queue.EnqueueN(data, len(data))
}

func (bt *TBluetooth) processQueue() {
	for bt.queue.GetDataLen() > 0 {
		data := bt.queue.PeekAll()
		_, packLen, removeLen, pack := bt.pickerFn(data, bt.queue.GetDataLen())

		if packLen > 0 {
			select {
			case bt.recvCh <- pack:
			default:
				log.Log.Error("Virtual BT recvCh full")
			}
		} else {
			break
		}

		if removeLen > 0 {
			bt.queue.DequeueN(int(removeLen))
		} else {
			break
		}
	}
}

// VirtualSerialRead 将来自 Flutter 的数据推入蓝牙接收队列
func (bt *TBluetooth) VirtualSerialRead(base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Log.Errorf("BT: Base64 解码失败: %v", err)
		return err
	}
	
	log.Log.Debugf("BT: 收到来自 Flutter 的数据 - 长度: %d 字节", len(data))
	
	select {
	case bt.notificationChan <- data:
		return nil
	default:
		log.Log.Error("BT: notification channel full, 丢弃数据")
		return fmt.Errorf("notification channel channel full")
	}
}
