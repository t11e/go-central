package central

import (
	"time"
)

type Role string

const (
	RoleNone    Role = ""
	RoleAdmin   Role = "admin"
	RolePartner Role = "partner"
	RoleUser    Role = "user"
)

type Organization struct {
	ID            int            `json:"id"`
	ParentID      *int           `json:"parent_id"`
	Title         string         `json:"title"`
	Path          string         `json:"path"`
	Realm         string         `json:"realm"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Organizations []Organization `json:"organizations"`
}

type Application struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	WriteAccess  bool          `json:"write_access"`
	CreatedAt    *time.Time    `json:"created_at"`
	UpdatedAt    *time.Time    `json:"updated_at"`
	Organization *Organization `json:"organization"`
}

type User struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IdentityID int       `json:"identity_id"`
	Admin      bool      `json:"admin"`
}

type Membership struct {
	ID             int           `json:"id"`
	Role           Role          `json:"role"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	OrganizationID int           `json:"organization_id"`
	User           *User         `json:"user"`
	Organization   *Organization `json:"organization"`
}
