package svc

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hirochachacha/go-smb2"
)

type SMBUploader struct {
	ServerIP  string
	ShareName string
	Username  string
	Password  string
}

func NewSMBUploader(serverIP, shareName, username, password string) *SMBUploader {
	return &SMBUploader{
		ServerIP:  serverIP,
		ShareName: shareName,
		Username:  username,
		Password:  password,
	}
}

// UploadFile 上传文件到共享文件夹
func (u *SMBUploader) UploadFile(localPath, remoteFileName string) (string, error) {
	// 检查本地文件
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return "stat local file failed", fmt.Errorf("stat local file %s failed: %v", localPath, err)
	}

	fmt.Printf("准备上传文件: %s (大小: %d bytes)\n",
		filepath.Base(localPath), fileInfo.Size())
	// 1. 建立TCP连接
	conn, err := net.DialTimeout("tcp", u.ServerIP+":445", 10*time.Second)
	if err != nil {
		return "connect to server failed", fmt.Errorf("connect to server %s failed: %v", u.ServerIP, err)
	}
	defer conn.Close()

	// 2. 创建SMB会话
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     u.Username,
			Password: u.Password,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return "smb handshake failed", fmt.Errorf("smb handshake failed: %v", err)
	}
	defer s.Logoff()

	// 3. 连接到共享文件夹
	fs, err := s.Mount(u.ShareName)
	if err != nil {
		return "mount share failed", fmt.Errorf("mount share %s failed: %v", u.ShareName, err)
	}
	defer fs.Umount()

	// 4. 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return "open local file failed", fmt.Errorf("open local file %s failed: %v", localPath, err)
	}
	defer localFile.Close()

	// 5. 创建远程文件
	remoteFile, err := fs.Create(remoteFileName)
	if err != nil {
		return "create remote file failed", fmt.Errorf("create remote file %s failed: %v", remoteFileName, err)
	}
	defer remoteFile.Close()

	// 6. 复制文件内容
	written, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return "copy file failed", fmt.Errorf("copy file %s to %s failed: %v", localPath, remoteFileName, err)
	}

	res := fmt.Sprintf("Successfully uploaded: %s (%d bytes)", remoteFileName, written)
	return res, nil

}

// ListFiles 列出共享文件夹中的文件
func (u *SMBUploader) ListFiles() error {
	conn, err := net.DialTimeout("tcp", u.ServerIP+":445", 10*time.Second)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     u.Username,
			Password: u.Password,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return fmt.Errorf("SMB握手失败: %v", err)
	}
	defer s.Logoff()

	fs, err := s.Mount(u.ShareName)
	if err != nil {
		return fmt.Errorf("挂载共享文件夹失败: %v", err)
	}
	defer fs.Umount()

	fmt.Printf("共享文件夹 '%s' 中的文件:\n", u.ShareName)
	fmt.Println("==================================")

	// 列出根目录文件
	infos, err := fs.ReadDir("")
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, info := range infos {
		name := info.Name()
		size := info.Size()
		modTime := info.ModTime().Format("2006-01-02 15:04:05")

		if info.IsDir() {
			fmt.Printf("[目录] %-40s %s\n", name, modTime)
		} else {
			fmt.Printf("[文件] %-40s %10d bytes %s\n", name, size, modTime)
		}
	}

	return nil
}

// func demo() {
// 	// 配置参数
// 	config := SMBUploader{
// 		ServerIP:  "10.5.52.62",
// 		ShareName: "csv文件",
// 		Username:  "FLF",
// 		Password:  "FLF111",
// 	}

// 	uploader := NewSMBUploader(
// 		config.ServerIP,
// 		config.ShareName,
// 		config.Username,
// 		config.Password,
// 	)

// 	// 显示菜单
// 	for {
// 		fmt.Println("\n=== CSV文件上传工具 ===")
// 		fmt.Println("1. 上传文件")
// 		fmt.Println("2. 查看共享文件夹内容")
// 		fmt.Println("3. 批量上传文件夹")
// 		fmt.Println("0. 退出")
// 		fmt.Print("请选择操作: ")

// 		var choice int
// 		fmt.Scanln(&choice)

// 		switch choice {
// 		case 1:
// 			uploadFile(uploader)
// 		case 2:
// 			listFiles(uploader)
// 		case 3:
// 			uploadDirectory(uploader)
// 		case 0:
// 			fmt.Println("程序退出")
// 			return
// 		default:
// 			fmt.Println("无效选择")
// 		}
// 	}
// }

func UploadFileByPath(uploader *SMBUploader, localPath string) {

	// 自动生成远程文件名
	remoteFileName := filepath.Base(localPath)

	fmt.Printf("正在上传 %s -> %s\n", localPath, remoteFileName)

	if res, err := uploader.UploadFile(localPath, remoteFileName); err != nil {
		log.Printf("上传失败: %v\n", err)
		UploadFmaWgtRecCsvLog(uploader.ServerIP, uploader.ShareName, res)
	} else {

		UploadFmaWgtRecCsvLog(uploader.ServerIP, uploader.ShareName, res)
	}

	SafeRemoveFile(localPath)
}

func SafeRemoveFile(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // 文件不存在，不需要删除
	}

	// 尝试删除文件
	if err := os.Remove(path); err != nil {
		// 特殊处理"文件不存在"错误（可能在其他地方被删除）
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("删除文件 %s 失败: %v", path, err)
	}

	// 验证文件是否真的被删除
	for i := 0; i < 3; i++ { // 重试机制
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil
		}
		time.Sleep(100 * time.Millisecond) // 等待一小段时间
	}

	// 文件仍然存在，删除可能失败
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("文件 %s 仍然存在，删除可能失败", path)
	}

	return nil
}

// func uploadFile(uploader *SMBUploader) {
// 	var localPath string
// 	fmt.Print("请输入本地文件路径: ")
// 	fmt.Scanln(&localPath)

// 	// 自动生成远程文件名
// 	remoteFileName := filepath.Base(localPath)

// 	fmt.Printf("正在上传 %s -> %s\n", localPath, remoteFileName)

// 	if res, err := uploader.UploadFile(localPath, remoteFileName); err != nil {
// 		log.Printf("上传失败: %v\n", err)
// 	} else {
// 		fmt.Println(res)
// 	}
// }

// func listFiles(uploader *SMBUploader) {
// 	if err := uploader.ListFiles(); err != nil {
// 		log.Printf("列出文件失败: %v\n", err)
// 	}
// }

// func uploadDirectory(uploader *SMBUploader) {
// 	var dirPath string
// 	fmt.Print("请输入本地文件夹路径: ")
// 	fmt.Scanln(&dirPath)

// 	// 遍历文件夹中的所有CSV文件
// 	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if !info.IsDir() && filepath.Ext(info.Name()) == ".csv" {
// 			remoteFileName := info.Name()
// 			fmt.Printf("上传: %s -> %s\n", path, remoteFileName)

// 			if res, err := uploader.UploadFile(path, remoteFileName); err != nil {
// 				log.Printf("上传失败 %s: %v\n", path, err)
// 			} else {
// 				fmt.Printf("✓ 上传成功: %s\n", res)
// 			}
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		log.Printf("遍历文件夹失败: %v\n", err)
// 	}
// }
