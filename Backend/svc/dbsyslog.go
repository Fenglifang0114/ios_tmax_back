package svc

import (
	"time"
	l "tmaxsrv/log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbSyslogRec struct {
	dbName string
}

func NewDbSyslogRec(dbName string) (*DbSyslogRec, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	if err = db.AutoMigrate(
		&Syslog{},
		&CalibrationLog{},
		&ScaleWgtLog{},
		&SealLog{},
	); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	return &DbSyslogRec{dbName: dbName}, nil
}

// 操作日志表
type Syslog struct {
	RecId         int       `gorm:"primaryKey;autoincrement;not null"` // 记录ID
	Operator      string    // 操作员
	RoleId        int       // 角色ID
	Module        string    // 模块  系统，app
	FuncName      string    // 功能模块  新增配方等具体功能
	OperationType string    // 操作类型  新增，删除，修改，查询等
	Operation     string    // 具体操作  新增了哪个配方，删除了哪个记录等
	Result        string    // 操作结果  成功，失败
	Remarks       string    // 备注
	CreateTime    time.Time // 创建时间
}

// 标定日志
// 由于多点标定时，中间不会校准，因此没有标定后的值
// 因此，不管多点还是单点，都记录一个点的误差值。
// 单点记录第一个，多点记录最后一个
type CalibrationLog struct {
	RecId      int       `gorm:"primaryKey;autoincrement;not null"` // 记录ID
	Operator   string    // 操作员
	RoleId     int       // 角色ID
	ScaleId    int       // 秤ID
	ScaleName  string    // 秤名称
	ModelName  string    // 秤机种
	Sn         string    // 秤号
	Type       string    // 标定类型  单点，线性
	Mode       string    // 标定单点，2点，3点
	Unit       string    // 标定单位
	Value      string    // 标定值
	Before     string    // 标定前值
	After      string    // 标定后值
	Error      string    // 标定误差
	Result     string    // 标定结果  成功，失败
	Remarks    string    // 备注
	CreateTime time.Time // 创建时间
}

// 称重日志
type ScaleWgtLog struct {
	RecId      int       `gorm:"primaryKey;autoincrement;not null"` // 记录ID
	Operator   string    // 操作员
	RoleId     int       // 角色ID
	Module     string    // app
	ScaleName  string    // 秤名称
	ModelName  string    // 秤机种
	Sn         string    // 秤号
	Weight     string    // 称重值
	Unit       string    // 单位
	Remarks    string    // 备注
	CreateTime time.Time // 创建时间
}

// 创建系统操作日志
func (d *DbSyslogRec) CreateSyslog(log Syslog) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Create(&log).Error; err != nil {
		l.Log.Debug("failed to create syslog")
	}
	return nil
}

// 创建标定日志
func (d *DbSyslogRec) CreateCalibrationLog(log CalibrationLog) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Create(&log).Error; err != nil {
		l.Log.Debug("failed to create calibration log")
	}
	return nil
}

// 创建称重日志
func (d *DbSyslogRec) CreateScaleLog(log ScaleWgtLog) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Create(&log).Error; err != nil {
		l.Log.Debug("failed to create scale log")
	}
	return nil
}

// 删除系统操作日志
func (d *DbSyslogRec) DeleteSyslog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&Syslog{}, "rec_id IN ?", recIds).Error; err != nil {
		l.Log.Debug("failed to delete syslog")
	}
	return nil
}

// 删除所有系统操作日志
func (d *DbSyslogRec) DeleteAllSyslog() error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
		return err // 返回错误
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
		return err // 返回错误
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	db.Where("rec_id > ?", 0).Delete(&Syslog{})
	return nil
}

// 删除标定日志
func (d *DbSyslogRec) DeleteCalibrationLog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&CalibrationLog{}, "rec_id IN ?", recIds).Error; err != nil {
		l.Log.Debug("failed to delete calibration log")
	}
	return nil
}

// 删除所有标定日志
func (d *DbSyslogRec) DeleteAllCalibrationLog() error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&CalibrationLog{}).Error; err != nil {
		l.Log.Debug("failed to delete all calibration log")
	}
	return nil
}

// 删除称重日志
func (d *DbSyslogRec) DeleteScaleLog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&ScaleWgtLog{}, "rec_id IN ?", recIds).Error; err != nil {
		l.Log.Debug("failed to delete scale log")
	}
	return nil
}

// 删除所有称重日志
func (d *DbSyslogRec) DeleteAllScaleLog() error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	db.Where("rec_id > ?", 0).Delete(&ScaleWgtLog{})

	return nil
}

// 删除多条系统操作日志
func (d *DbSyslogRec) DeleteMultiSyslog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&Syslog{}, recIds).Error; err != nil {
		l.Log.Debug("failed to delete multi syslog")
	}
	return nil
}

// 删除多条标定日志
func (d *DbSyslogRec) DeleteMultiCalibrationLog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&CalibrationLog{}, recIds).Error; err != nil {
		l.Log.Debug("failed to delete multi calibration log")
	}
	return nil
}

// 删除多条称重日志
func (d *DbSyslogRec) DeleteMultiScaleLog(recIds []int) error {
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	if err = db.Delete(&ScaleWgtLog{}, recIds).Error; err != nil {
		l.Log.Debug("failed to delete multi scale log")
	}
	return nil
}

// 获取所有系统操作日志
func (d *DbSyslogRec) GetAllSyslog() ([]Syslog, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	var syslogs []Syslog
	db.Find(&syslogs)
	return syslogs, nil
}

// 获取所有标定日志
func (d *DbSyslogRec) GetAllCalibrationLog() ([]CalibrationLog, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	var calibrationLogs []CalibrationLog
	db.Find(&calibrationLogs)
	return calibrationLogs, nil
}

// 获取所有称重日志
func (d *DbSyslogRec) GetAllScaleLog() ([]ScaleWgtLog, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	var scaleLogs []ScaleWgtLog
	db.Find(&scaleLogs)
	return scaleLogs, nil
}

// 定义查询条件结构体
type SyslogQuery struct {
	Operator  string // 操作员
	Module    string
	RoleId    int       // 角色ID
	StartTime time.Time // 创建时间起始
	EndTime   time.Time // 创建时间结束
}

// 分页获取系统操作日志，带有排序字段，带排序方向，页码，页大小，查询条件
// 返回列表和查询的总数
// 分页获取系统操作日志（带组合条件、排序、分页）
func (d *DbSyslogRec) GetSyslogWithSort(
	pageNum int,
	pageSize int,
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]Syslog, int64, error) {
	if pageNum <= 0 {
		return []Syslog{}, 0, nil
	}
	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, 0, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, 0, err
	}
	defer sqlDB.Close() // 确保连接关闭

	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&Syslog{}) // 基于Syslog表构建查询

	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}
	if query.Module != "" {
		tx = tx.Where("module LIKE ?", "%"+query.Module+"%") // 模糊匹配模块
	}
	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}

	// 3. 统计符合条件的总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		l.Log.Debug("failed to count syslogs:", err)
		return nil, 0, err
	}

	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder

	// 5. 分页查询符合条件的数据
	var syslogs []Syslog
	if err := tx.
		Order(orderStr).
		Limit(pageSize).
		Offset((pageNum - 1) * pageSize). // 计算偏移量（注意pageNum=0的边界处理）
		Find(&syslogs).Error; err != nil {
		l.Log.Debug("failed to query syslogs:", err)
		return nil, 0, err
	}

	return syslogs, total, nil
}

// 分页获取标定日志（带组合条件、排序、分页）
// 返回列表和查询的总数
func (d *DbSyslogRec) GetCalibrationLogWithSort(
	pageNum int,
	pageSize int,
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]CalibrationLog, int64, error) {
	if pageNum <= 0 {
		return []CalibrationLog{}, 0, nil
	}
	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, 0, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, 0, err
	}
	defer sqlDB.Close() // 确保连接关闭

	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&CalibrationLog{}) // 基于CalibrationLog表构建查询

	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}

	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}
	// 3. 统计符合条件的总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		l.Log.Debug("failed to count calibration logs:", err)
		return nil, 0, err
	}
	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder
	// 5. 分页查询符合条件的数据
	var calibrationLogs []CalibrationLog
	if err := tx.
		Order(orderStr).
		Limit(pageSize).
		Offset((pageNum - 1) * pageSize). // 计算偏移量（注意pageNum=0的边界处理）
		Find(&calibrationLogs).Error; err != nil {
		l.Log.Debug("failed to query calibration logs:", err)
		return nil, 0, err
	}
	return calibrationLogs, total, nil
}

// 分页获取称重日志（带组合条件、排序、分页）
// 返回列表和查询的总数
func (d *DbSyslogRec) GetScaleLogWithSort(
	pageNum int,
	pageSize int,
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]ScaleWgtLog, int64, error) {

	if pageNum <= 0 {
		return []ScaleWgtLog{}, 0, nil
	}

	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, 0, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, 0, err
	}
	defer sqlDB.Close() // 确保连接关闭
	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&ScaleWgtLog{}) // 基于WeightLog表构建查询
	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}

	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}
	// 3. 统计符合条件的总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		l.Log.Debug("failed to count weight logs:", err)
		return nil, 0, err
	}
	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder
	// 5. 分页查询符合条件的数据
	var weightLogs []ScaleWgtLog
	if err := tx.
		Order(orderStr).
		Limit(pageSize).
		Offset((pageNum - 1) * pageSize). // 计算偏移量（注意pageNum=0的边界处理）
		Find(&weightLogs).Error; err != nil {
		l.Log.Debug("failed to query weight logs:", err)
		return nil, 0, err
	}
	return weightLogs, total, nil
}

// 分页获取系统操作日志，带有排序字段，带排序方向，页码，页大小，查询条件
// 返回列表和查询的总数
// 分页获取系统操作日志（带组合条件、排序、分页）
func (d *DbSyslogRec) ExportSyslogWithSort(
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]Syslog, error) {

	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close() // 确保连接关闭

	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&Syslog{}) // 基于Syslog表构建查询

	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}
	if query.Module != "" {
		tx = tx.Where("module LIKE ?", "%"+query.Module+"%") // 模糊匹配模块
	}
	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}

	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder

	// 5. 分页查询符合条件的数据
	var syslogs []Syslog
	if err := tx.
		Order(orderStr).
		Find(&syslogs).Error; err != nil {
		l.Log.Debug("failed to query syslogs:", err)
		return nil, err
	}
	return syslogs, nil
}

// 分页获取标定日志（带组合条件、排序、分页）
// 返回列表和查询的总数
func (d *DbSyslogRec) ExportCalibrationLogWithSort(
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]CalibrationLog, error) {

	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close() // 确保连接关闭

	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&CalibrationLog{}) // 基于CalibrationLog表构建查询

	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}

	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}

	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder
	// 5. 分页查询符合条件的数据
	var calibrationLogs []CalibrationLog
	if err := tx.
		Order(orderStr).
		Find(&calibrationLogs).Error; err != nil {
		l.Log.Debug("failed to query calibration logs:", err)
		return nil, err
	}
	return calibrationLogs, nil
}

// 分页获取称重日志（带组合条件、排序、分页）
// 返回列表和查询的总数
func (d *DbSyslogRec) ExportScaleLogWithSort(
	sortField string,
	sortOrder string,
	query SyslogQuery,
) ([]ScaleWgtLog, error) {

	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close() // 确保连接关闭
	// 2. 构建查询条件（基于空Model开始，避免直接操作db导致全局污染）
	tx := db.Model(&ScaleWgtLog{}) // 基于WeightLog表构建查询
	// 动态添加条件（仅当条件非空时生效）
	if query.Operator != "" {
		tx = tx.Where("operator LIKE ?", "%"+query.Operator+"%") // 模糊匹配操作员
	}

	if query.RoleId != 0 { // 假设RoleId为0时表示无此条件
		tx = tx.Where("role_id = ?", query.RoleId) // 精确匹配角色ID
	}
	if !query.StartTime.IsZero() {
		tx = tx.Where("create_time >= ?", query.StartTime) // 大于等于起始时间
	}
	if !query.EndTime.IsZero() {
		tx = tx.Where("create_time <= ?", query.EndTime) // 小于等于结束时间
	}
	// 3. 统计符合条件的总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		l.Log.Debug("failed to count weight logs:", err)
		return nil, err
	}
	// 4. 处理排序（默认按创建时间倒序，避免无效排序）
	if sortField == "" {
		sortField = "create_time" // 默认排序字段
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc" // 默认倒序
	}
	orderStr := sortField + " " + sortOrder
	// 5. 分页查询符合条件的数据
	var weightLogs []ScaleWgtLog
	if err := tx.
		Order(orderStr).
		Find(&weightLogs).Error; err != nil {
		l.Log.Debug("failed to query weight logs:", err)
		return nil, err
	}
	return weightLogs, nil
}

// 根据日志ID列表获取系统日志记录
func (d *DbSyslogRec) GetSyslogByID(RecIds []int) ([]Syslog, error) {
	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close()

	tx := db.Model(&Syslog{})

	var syslogs []Syslog
	if err := tx.
		Where("rec_id IN ?", RecIds).
		Find(&syslogs).Error; err != nil {
		l.Log.Debug("failed to query syslog by rec_id:", err)
		return nil, err
	}
	return syslogs, nil
}

// 根据日志ID列表获取校准日志记录
func (d *DbSyslogRec) GetCalibrationLogByID(RecIds []int) ([]CalibrationLog, error) {
	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close()

	tx := db.Model(&CalibrationLog{})

	var calibrationLogs []CalibrationLog
	if err := tx.
		Where("rec_id IN ?", RecIds).
		Find(&calibrationLogs).Error; err != nil {
		l.Log.Debug("failed to query calibration log by rec_id:", err)
		return nil, err
	}
	return calibrationLogs, nil
}

// 根据日志ID列表获取称重日志记录
func (d *DbSyslogRec) GetScaleLogByID(RecIds []int) ([]ScaleWgtLog, error) {
	// 1. 连接数据库（建议抽离为全局连接，避免重复打开关闭）
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database:", err)
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to get DB instance:", err)
		return nil, err
	}
	defer sqlDB.Close()

	tx := db.Model(&ScaleWgtLog{})

	var scaleLogs []ScaleWgtLog
	if err := tx.
		Where("rec_id IN ?", RecIds).
		Find(&scaleLogs).Error; err != nil {
		l.Log.Debug("failed to query scale log by rec_id:", err)
		return nil, err
	}
	return scaleLogs, nil
}

type SealLog struct {
	RecId         int `gorm:"primaryKey;not null;autoincrement;"`
	Model         string
	ScaleId       int
	Sn            string
	RoleId        int `gorm:"not null;"`
	Operator      string
	Operation     string    //seal unseal
	OperationTime time.Time `gorm:"autoCreateTime"`
	Remark        string
	Result        string
}

func (d *DbSyslogRec) CreateSealLog(log *SealLog) error {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	if err = db.Create(log).Error; err != nil {
		return err
	}
	return nil
}

// 查询印章日志列表
func (d *DbSyslogRec) ListSealLogs(model string, sn string) ([]SealLog, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return []SealLog{}, err

	}
	sqlDB, err := db.DB()
	if err != nil {
		return []SealLog{}, err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	var logs []SealLog

	result := db.Where("model = ? and sn = ?", model, sn).Find(&logs)
	if result.Error != nil {
		return []SealLog{}, err
	}
	if result.RowsAffected == 0 {
		return []SealLog{}, err
	}
	return logs, result.Error
}
