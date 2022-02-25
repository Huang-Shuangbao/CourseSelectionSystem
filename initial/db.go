package initial

import (
	"camp-backend/types"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dbUser     = "root"
	dbPassword = "bytedancecamp"
	dbServer   = "180.184.68.166"
	dbName     = "camp"
)

var Db *gorm.DB

func SetupDatasource() {
	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbServer + ":3306)/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	Db, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Sprintf("open mysql failed, err is %s", err))
	}

	Db.AutoMigrate(&types.TMember{})
	Db.AutoMigrate(&types.TCourse{})
}
