package svc

var portsListed PortsListed

type PortsListed struct {
	handlers []interface{ Handle() }
}

// Register adds an event handler for this event
func (u *PortsListed) Register(handler interface{ Handle() }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u PortsListed) Trigger() {
	for _, handler := range u.handlers {
		go handler.Handle()
	}
}

var scalesListedSrv ScaleListedSrv

type ScaleListedSrv struct {
	handlers []interface {
		Handle(scaleMgr *ScaleMgr, scaleId int64)
	}
}

// Register adds an event handler for this event
func (u *ScaleListedSrv) Register(handler interface {
	Handle(payload *ScaleMgr, scaleId int64)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleListedSrv) Trigger(scaleMgr *ScaleMgr, scaleId int64) {
	for _, handler := range u.handlers {
		go handler.Handle(scaleMgr, scaleId)
	}
}

var scalesListed ScaleListed

type ScaleListed struct {
	handlers []interface{ Handle(scaleMgr *ScaleMgr) }
}

// Register adds an event handler for this event
func (u *ScaleListed) Register(handler interface{ Handle(payload *ScaleMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleListed) Trigger(payload *ScaleMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var scaleAdded ScaleAdded

type ScaleAdded struct {
	handlers []interface{ Handle(payload ReqAddScale) }
}

// Register adds an event handler for this event
func (u *ScaleAdded) Register(handler interface{ Handle(ReqAddScale) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleAdded) Trigger(payload ReqAddScale) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var scaleModified ScaleModified

type ScaleModified struct {
	handlers []interface{ Handle(payload ReqModifyScale) }
}

// Register adds an event handler for this event
func (u *ScaleModified) Register(handler interface{ Handle(payload ReqModifyScale) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleModified) Trigger(payload ReqModifyScale) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var scaleNameModified ScaleNameModified

type ScaleNameModified struct {
	handlers []interface {
		Handle(payload ReqModifyScaleName)
	}
}

// Register modify an event handler for this event
func (u *ScaleNameModified) Register(handler interface {
	Handle(payload ReqModifyScaleName)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleNameModified) Trigger(payload ReqModifyScaleName) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var scaleDeleted ScaleDeleted

type ScaleDeleted struct {
	handlers []interface{ Handle(payload ReqDelScale) }
}

// Register adds an event handler for this event
func (u *ScaleDeleted) Register(handler interface{ Handle(payload ReqDelScale) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleDeleted) Trigger(payload ReqDelScale) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var productsListed ProductListed

type ProductListed struct {
	handlers []interface{ Handle(mgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *ProductListed) Register(handler interface{ Handle(mgr *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductListed) Trigger(mgr *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr)
	}
}

var productAdded ProductAdded

type ProductAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddProductList)
	}
}

// Register adds an event handler for this event
func (u *ProductAdded) Register(handler interface {
	Handle(*SrvMgr, ReqAddProductList)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductAdded) Trigger(mgr *SrvMgr, payload ReqAddProductList) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var productModified ProductModified

type ProductModified struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddProductList)
	}
}

// Register adds an event handler for this event
func (u *ProductModified) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqAddProductList)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductModified) Trigger(mgr *SrvMgr, payload ReqAddProductList) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var productDeleted ProductDeleted

type ProductDeleted struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqDelProduct)
	}
}

// Register adds an event handler for this event
func (u *ProductDeleted) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqDelProduct)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductDeleted) Trigger(mgr *SrvMgr, payload ReqDelProduct) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var productDeletedAll ProductDeletedAll

type ProductDeletedAll struct {
	handlers []interface {
		Handle(mgr *SrvMgr)
	}
}

// Register adds an event handler for this event

func (u *ProductDeletedAll) Register(handler interface {
	Handle(mgr *SrvMgr)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductDeletedAll) Trigger(mgr *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr)
	}
}

var usersListed UserListed

type UserListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *UserListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u UserListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var userAdded UserAdded

type UserAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddUser)
	}
}

// Register adds an event handler for this event
func (u *UserAdded) Register(handler interface{ Handle(*SrvMgr, ReqAddUser) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u UserAdded) Trigger(mgr *SrvMgr, payload ReqAddUser) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var userModified UserModified

type UserModified struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqModifyUser)
	}
}

// Register adds an event handler for this event
func (u *UserModified) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqModifyUser)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u UserModified) Trigger(mgr *SrvMgr, payload ReqModifyUser) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var userDeleted UserDeleted

type UserDeleted struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqDelUser)
	}
}

// Register adds an event handler for this event
func (u *UserDeleted) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqDelUser)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u UserDeleted) Trigger(mgr *SrvMgr, payload ReqDelUser) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var detailListed DetailListed

type DetailListed struct {
	handlers []interface{ Handle(scaleMgr *ScaleMgr) }
}

// Register adds an event handler for this event
func (u *DetailListed) Register(handler interface{ Handle(payload *ScaleMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u DetailListed) Trigger(payload *ScaleMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var scaleSrvList ScaleSrvList

type ScaleSrvList struct {
	handlers []interface {
		Handle(scaleMgr *ScaleMgr, scaleIdStr string)
	}
}

// Register adds an event handler for this event
func (u *ScaleSrvList) Register(handler interface {
	Handle(payload *ScaleMgr, scaleIdStr string)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ScaleSrvList) Trigger(payload *ScaleMgr, scaleIdStr string) {
	for _, handler := range u.handlers {
		go handler.Handle(payload, scaleIdStr)
	}
}

var setScaleSrvVal SetScaleSrvVal

type SetScaleSrvVal struct {
	handlers []interface {
		Handle(scaleMgr *ScaleMgr, relInfo SrvScaleRel)
	}
}

// Register adds an event handler for this event
func (u *SetScaleSrvVal) Register(handler interface {
	Handle(payload *ScaleMgr, relInfo SrvScaleRel)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u SetScaleSrvVal) Trigger(payload *ScaleMgr, relInfo SrvScaleRel) {
	for _, handler := range u.handlers {
		go handler.Handle(payload, relInfo)
	}
}

var wifiListed WifiListed

type WifiListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *WifiListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u WifiListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

var wifiAdded WifiAdded

type WifiAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddWifi)
	}
}

// Register adds an event handler for this event
func (u *WifiAdded) Register(handler interface{ Handle(*SrvMgr, ReqAddWifi) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u WifiAdded) Trigger(mgr *SrvMgr, payload ReqAddWifi) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var sendToSrv1 SendToSrv1

type SendToSrv1 struct {
	handlers []interface {
		Handle(mgr *SrvMgr, jsonStr string)
	}
}

// Register adds an event handler for this event
func (u *SendToSrv1) Register(handler interface {
	Handle(payload *SrvMgr, jsonStr string)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u SendToSrv1) Trigger(mgr *SrvMgr, jsonStr string) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, jsonStr)
	}
}

var sendToUi SendToUi

type SendToUi struct {
	handlers []interface {
		Handle(mgr *SrvMgr, jsonStr string)
	}
}

// Register adds an event handler for this event
func (u *SendToUi) Register(handler interface {
	Handle(payload *SrvMgr, jsonStr string)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u SendToUi) Trigger(mgr *SrvMgr, jsonStr string) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, jsonStr)
	}
}

var doServiceAction DoServiceAction

type DoServiceAction struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqDoServiceAction)
	}
}

// Register adds an event handler for this event
func (u *DoServiceAction) Register(handler interface {
	Handle(*SrvMgr, ReqDoServiceAction)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u DoServiceAction) Trigger(mgr *SrvMgr, payload ReqDoServiceAction) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 配方秤
// 新增原料类型
var rawTypeAdded RawTypeAdded

type RawTypeAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddRawType)
	}
}

// Register adds an event handler for this event
func (u *RawTypeAdded) Register(handler interface{ Handle(*SrvMgr, ReqAddRawType) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawTypeAdded) Trigger(mgr *SrvMgr, payload ReqAddRawType) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 新增配方类型
var formulaTypeAdded FormulaTypeAdded

type FormulaTypeAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddFormulaType)
	}
}

// Register adds an event handler for this event
func (u *FormulaTypeAdded) Register(handler interface {
	Handle(*SrvMgr, ReqAddFormulaType)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaTypeAdded) Trigger(mgr *SrvMgr, payload ReqAddFormulaType) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

//获取配方类型列表

var formulaTypeListed FormulaTypeListed

type FormulaTypeListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *FormulaTypeListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaTypeListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

//获取配方类型列表

var rawTypeListed RawTypeListed

type RawTypeListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *RawTypeListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawTypeListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

// 新增原料数据
var rawDataAdded RawDataAdded

type RawDataAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddRawData)
	}
}

// Register adds an event handler for this event
func (u *RawDataAdded) Register(handler interface{ Handle(*SrvMgr, ReqAddRawData) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawDataAdded) Trigger(mgr *SrvMgr, payload ReqAddRawData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 获取原料数据列表
var rawDataListed RawDataListed

type RawDataListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *RawDataListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawDataListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

// 修改原料数据
var rawDataEdited RawDataEdited

type RawDataEdited struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqEditRawData)
	}
}

// Register adds an event handler for this event
func (u *RawDataEdited) Register(handler interface{ Handle(*SrvMgr, ReqEditRawData) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawDataEdited) Trigger(mgr *SrvMgr, payload ReqEditRawData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 删除原料数据
var rawDataDeleted RawDataDeleted

type RawDataDeleted struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqDelRawData)
	}
}

// Register adds an event handler for this event
func (u *RawDataDeleted) Register(handler interface{ Handle(*SrvMgr, ReqDelRawData) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u RawDataDeleted) Trigger(mgr *SrvMgr, payload ReqDelRawData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var formulaDataAdded FormulaDataAdded

type FormulaDataAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddFormulaData)
	}
}

// Register adds an event handler for this event
func (u *FormulaDataAdded) Register(handler interface {
	Handle(*SrvMgr, ReqAddFormulaData)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaDataAdded) Trigger(mgr *SrvMgr, payload ReqAddFormulaData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var formulaDataEdited FormulaDataEdited

type FormulaDataEdited struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqAddFormulaData)
	}
}

// Register adds an event handler for this event
func (u *FormulaDataEdited) Register(handler interface {
	Handle(*SrvMgr, ReqAddFormulaData)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaDataEdited) Trigger(mgr *SrvMgr, payload ReqAddFormulaData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var formulaRecList FormulaRecListed

type FormulaRecListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *FormulaRecListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaRecListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

// 新增配方称重记录
var formulaWgtRecAdded FormulaWgtRecAdded

type FormulaWgtRecAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqFormulaWgtRec)
	}
}

// Register adds an event handler for this event
func (u *FormulaWgtRecAdded) Register(handler interface {
	Handle(*SrvMgr, ReqFormulaWgtRec)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaWgtRecAdded) Trigger(mgr *SrvMgr, payload ReqFormulaWgtRec) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 获取配方称重记录列表
var formulaWgtRecList FormulaWgtRecListed

type FormulaWgtRecListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *FormulaWgtRecListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaWgtRecListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}

// 根据RecId删除配方
var formulaDeleted FormulaDeleted

type FormulaDeleted struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqDelFmaData)
	}
}

// Register adds an event handler for this event
func (u *FormulaDeleted) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqDelFmaData)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FormulaDeleted) Trigger(mgr *SrvMgr, payload ReqDelFmaData) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 增加流速
var flowRateAdded FlowRateAdded

type FlowRateAdded struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqFlowRateRec)
	}
}

// Register adds an event handler for this event
func (u *FlowRateAdded) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqFlowRateRec)
}) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FlowRateAdded) Trigger(mgr *SrvMgr, payload ReqFlowRateRec) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

// 获取流速列表
var flowRateList FlowRateListed

type FlowRateListed struct {
	handlers []interface{ Handle(srvMgr *SrvMgr) }
}

// Register adds an event handler for this event
func (u *FlowRateListed) Register(handler interface{ Handle(payload *SrvMgr) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u FlowRateListed) Trigger(payload *SrvMgr) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
