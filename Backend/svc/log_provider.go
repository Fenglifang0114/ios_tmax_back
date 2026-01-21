package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type LogRecProvider struct {
	myId   string
	infoPb *DbSyslogRec
}

func NewLogRecProvider() *LogRecProvider {
	database := filepath.Join(comm.GetSrvDataPath(), "syslog.db")
	infoPb, _ := NewDbSyslogRec(database)
	return &LogRecProvider{myId: "LogRecProvider", infoPb: infoPb}
}

// 新增系统日志记录
func (p *LogRecProvider) AddSyslog(rec Syslog) error {
	return p.infoPb.CreateSyslog(rec)
}

// 新增校准日志记录
func (p *LogRecProvider) AddCalibrationLog(rec CalibrationLog) error {
	return p.infoPb.CreateCalibrationLog(rec)
}

// 新增秤日志记录
func (p *LogRecProvider) AddScaleLog(rec ScaleWgtLog) error {
	return p.infoPb.CreateScaleLog(rec)
}

// 删除系统日志记录
func (p *LogRecProvider) DeleteSyslog(RecId []int) error {
	return p.infoPb.DeleteSyslog(RecId)
}

// 删除校准日志记录
func (p *LogRecProvider) DeleteCalibrationLog(RecId []int) error {
	return p.infoPb.DeleteCalibrationLog(RecId)
}

// 删除称重日志记录
func (p *LogRecProvider) DeleteScaleLog(RecId []int) error {
	return p.infoPb.DeleteScaleLog(RecId)
}

// 删除所有系统日志记录
func (p *LogRecProvider) DeleteAllSyslog() error {
	return p.infoPb.DeleteAllSyslog()
}

// 删除所有校准日志记录
func (p *LogRecProvider) DeleteAllCalibrationLog() error {
	return p.infoPb.DeleteAllCalibrationLog()
}

// 删除所有称重日志记录
func (p *LogRecProvider) DeleteAllScaleLog() error {
	return p.infoPb.DeleteAllScaleLog()
}

// 获取系统日志记录
func (p *LogRecProvider) GetSyslog(page int, pageSize int, fieldName string, direction string, query SyslogQuery) ([]Syslog, int64, error) {
	return p.infoPb.GetSyslogWithSort(page, pageSize, fieldName, direction, query)
}

// 获取校准日志记录
func (p *LogRecProvider) GetCalibrationLog(page int, pageSize int, fieldName string, direction string, query SyslogQuery) ([]CalibrationLog, int64, error) {
	return p.infoPb.GetCalibrationLogWithSort(page, pageSize, fieldName, direction, query)
}

// 获取称重日志记录
func (p *LogRecProvider) GetScaleLog(page int, pageSize int, fieldName string, direction string, query SyslogQuery) ([]ScaleWgtLog, int64, error) {
	return p.infoPb.GetScaleLogWithSort(page, pageSize, fieldName, direction, query)
}

// 获取所有称重日志记录
func (p *LogRecProvider) GetAllScaleLog() ([]ScaleWgtLog, error) {
	return p.infoPb.GetAllScaleLog()
}

// 获取所有系统日志记录
func (p *LogRecProvider) GetAllSyslog() ([]Syslog, error) {
	return p.infoPb.GetAllSyslog()
}

// 获取所有校准日志记录
func (p *LogRecProvider) GetAllCalibrationLog() ([]CalibrationLog, error) {
	return p.infoPb.GetAllCalibrationLog()
}

// 导出满足条件的系统日志记录
func (p *LogRecProvider) ExportSyslog(fieldName string, direction string, query SyslogQuery) ([]Syslog, error) {
	return p.infoPb.ExportSyslogWithSort(fieldName, direction, query)
}

// 导出满足条件的校准日志记录
func (p *LogRecProvider) ExportCalibrationLog(fieldName string, direction string, query SyslogQuery) ([]CalibrationLog, error) {
	return p.infoPb.ExportCalibrationLogWithSort(fieldName, direction, query)
}

// 导出满足条件的称重日志记录
func (p *LogRecProvider) ExportScaleLog(fieldName string, direction string, query SyslogQuery) ([]ScaleWgtLog, error) {
	return p.infoPb.ExportScaleLogWithSort(fieldName, direction, query)
}

// 根据日志ID列表获取系统日志记录
func (p *LogRecProvider) GetSyslogByID(RecIds []int) ([]Syslog, error) {
	return p.infoPb.GetSyslogByID(RecIds)
}

// 根据日志ID列表获取校准日志记录
func (p *LogRecProvider) GetCalibrationLogByID(RecIds []int) ([]CalibrationLog, error) {
	return p.infoPb.GetCalibrationLogByID(RecIds)
}

// 根据日志ID列表获取称重日志记录
func (p *LogRecProvider) GetScaleLogByID(RecIds []int) ([]ScaleWgtLog, error) {
	return p.infoPb.GetScaleLogByID(RecIds)
}

// 获取所有铅封日志记录
func (p *LogRecProvider) GetAllSealLog(model string, sn string) ([]SealLog, error) {
	return p.infoPb.ListSealLogs(model, sn)
}

// 创建铅封日志
func (p *LogRecProvider) AddSealLog(rec SealLog) error {
	return p.infoPb.CreateSealLog(&rec)
}
