package types

// 系统内置管理员账号
// 账号名：JudgeAdmin 密码：JudgePassword2022

type TMember struct {
	UserID       string `gorm:"primaryKey"`
	Nickname     string
	Username     string    `gorm:"unique;not null"`
	Password     string    `json:"-" gorm:"not null"`
	UserType     UserType  `gorm:"not null"`
	IsDeleted    bool      `json:"-" gorm:"not null;default:false"`
	LearnCourses []TCourse `json:"-" gorm:"many2many:learn_courses;joinForeignKey:user_id;joinReferences:course_id;"`
}

type TCourse struct {
	CourseID  string  `gorm:"primaryKey"`
	Name      string  `gorm:"not null;unique"`
	TeacherID string  `gorm:"size:191"`
	Teacher   TMember `json:"-" gorm:"foreignKey:TeacherID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Capacity  int     `json:"-" gorm:"not null;default:0"`
}

type UserType int

const (
	Admin   UserType = 1
	Student UserType = 2
	Teacher UserType = 3
)

func (u UserType) String() string {
	switch u {
	case Admin:
		return "Admin"
	case Student:
		return "Student"
	case Teacher:
		return "Teacher"
	default:
		return "Unknown"
	}
}
