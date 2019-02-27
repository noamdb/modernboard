package controllers

import (
	"gitlab.com/noamdb/modernboard/utils"
)

type threadCreate struct {
	boardURI string
	subject  string
	author   string
	tripcode string
	body     string
	fileName string
}

func (tc threadCreate) valid() bool {
	return utils.ValidLength(tc.subject, -1, 100) &&
		utils.ValidLength(tc.author, 1, 20) &&
		utils.ValidLength(tc.tripcode, -1, 30) &&
		utils.ValidLength(tc.body, -1, 15000)
}

type PostCreate struct {
	threadID int
	author   string
	tripcode string
	body     string
	Bump     bool
}

func (pc PostCreate) valid() bool {
	return utils.ValidLength(pc.author, 1, 20) &&
		utils.ValidLength(pc.tripcode, 0, 30) &&
		utils.ValidLength(pc.body, 0, 15000)
}

type UserLogin struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (ul UserLogin) Valid() bool {
	return utils.ValidLength(ul.Name, 1, 20) &&
		utils.ValidLength(ul.Password, 4, 30)
}

type UserLoginResponse struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

type Report struct {
	Reason string `json:"reason"`
}

func (r Report) valid() bool {
	return utils.ValidLength(r.Reason, 1, 100)
}

type BanPosterCreate struct {
	Reason string `json:"reason"`
}

func (bpc BanPosterCreate) valid() bool {
	return utils.ValidLength(bpc.Reason, 1, 100)
}

type BanIPCreate struct {
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

func (bip BanIPCreate) valid() bool {
	return utils.ValidLength(bip.Reason, 1, 100) &&
		utils.ValidLength(bip.IP, 1, 100)
}

type ChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (o ChangePassword) valid() bool {
	return utils.ValidLength(o.OldPassword, 4, 100) &&
		utils.ValidLength(o.NewPassword, 4, 100)
}

type BoardUserCreate struct {
	Name string `json:"name"`
}

func (o BoardUserCreate) Valid() bool {
	return utils.ValidLength(o.Name, 1, 20)
}
