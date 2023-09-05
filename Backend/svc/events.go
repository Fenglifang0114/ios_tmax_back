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
		Handle(mgr *SrvMgr, payload ReqAddProduct)
	}
}

// Register adds an event handler for this event
func (u *ProductAdded) Register(handler interface{ Handle(*SrvMgr, ReqAddProduct) }) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductAdded) Trigger(mgr *SrvMgr, payload ReqAddProduct) {
	for _, handler := range u.handlers {
		go handler.Handle(mgr, payload)
	}
}

var productModified ProductModified

type ProductModified struct {
	handlers []interface {
		Handle(mgr *SrvMgr, payload ReqModifyProduct)
	}
}

// Register adds an event handler for this event
func (u *ProductModified) Register(handler interface {
	Handle(mgr *SrvMgr, payload ReqModifyProduct)
},
) {
	u.handlers = append(u.handlers, handler)
}

// Trigger sends out an event with the payload
func (u ProductModified) Trigger(mgr *SrvMgr, payload ReqModifyProduct) {
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
