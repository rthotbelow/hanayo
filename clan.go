package main

import (
	"database/sql"
	"strconv"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"time"
)

// TODO: replace with simple ResponseInfo containing userid
type clanData struct {
	baseTemplateData
	ClanID int
}


func leaveClan(c *gin.Context) {
	i := c.Param("cid")
	// login check
	if getContext(c).User.ID == 0 {
			resp403(c)
			return
		}
	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ? AND perms = 8", getContext(c).User.ID, i).
		Scan(new(int)) == sql.ErrNoRows {
		// ดูว่าคนนี้มีแคลนหรือยัง
		if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i).
			Scan(new(int)) == sql.ErrNoRows {
			addMessage(c, errorMessage{T(c, "นี้มันปัญหาอะไรวะเนี่ย...")})
			return
		}
		// กูไม่รู้หรอกว่ามันจะได้ผลมั้ย แต่ควยชั่งแม่งเย็ดแม่
		
			
		db.Exec("DELETE FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i)
		addMessage(c, successMessage{T(c, "ออกจากแคลนแล้ว")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+i)
	} else {
		// เดี๋ยวไอ้เหี้ย มันออกไปยังวะ!!!
		if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i).
			Scan(new(int)) == sql.ErrNoRows {
			addMessage(c, errorMessage{T(c, "นี้มันปัญหาอะไรวะเนี่ย...")})
			return
		}
		// ลบคำเชิญออก
		db.Exec("DELETE FROM clans_invites WHERE clan = ?", i)
		// ลบทุกคนออกจากแคลน :c
		db.Exec("DELETE FROM user_clans WHERE clan = ?", i)
		// ควยไม่สร้างแม่งละสัส :c
		db.Exec("DELETE FROM clans WHERE id = ?", i)
		
		addMessage(c, successMessage{T(c, "ยุบแคลนแล้ว")})
		getSession(c).Save()
		c.Redirect(302, "/clans?mode=0")
	}
	

}


func clanPage(c *gin.Context) {
	var (
		clanID           int
		clanName         string
		clanDescription  string
		clanIcon         string
	)

	// ctx := getContext(c)

	i := c.Param("cid")
	if _, err := strconv.Atoi(i); err != nil {
		err := db.QueryRow("SELECT id, name, description, icon FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := db.QueryRow(`SELECT id, name, description, icon FROM clans WHERE id = ? LIMIT 1`, i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
		switch {
		case err == nil:
		case err == sql.ErrNoRows:
			err := db.QueryRow("SELECT id, name, description, icon FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
			if err != nil && err != sql.ErrNoRows {
				c.Error(err)
			}
		default:
			c.Error(err)
		}
	}

	data := new(clanData)
	data.ClanID = clanID
	defer resp(c, 200, "clansample.html", data)

	if data.ClanID == 0 {
		data.TitleBar = "ไม่เจอแคลนนี้นะ"
		data.Messages = append(data.Messages, warningMessage{T(c, "That clan could not be found.")})
		return
	}

	if getContext(c).User.Privileges&1 > 0 {
		if db.QueryRow("SELECT 1 FROM clans WHERE clan = ?", clanID).Scan(new(string)) != sql.ErrNoRows {
			var bg string
			db.QueryRow("SELECT background FROM clans WHERE id = ?", clanID).Scan(&bg)
			data.KyutGrill = bg
			data.KyutGrillAbsolute = true
		}
	}

	data.TitleBar = T(c, "หน้าเว๊ปของแคลน %s", clanName)
	data.DisableHH = true
	// data.Scripts = append(data.Scripts, "/static/profile.js")
}

func checkCount(rows *sql.Rows) (count int) {
 	for rows.Next() {
    	err:= rows.Scan(&count)
    	if err != nil {
			panic(err)
		}
    }   
    return count
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano()+int64(3))
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func createInvite(c *gin.Context) {
ctx := getContext(c)
	if string(c.PostForm("description")) == "" && string(c.PostForm("icon")) == "" && string(c.PostForm("tag")) == "" && string(c.PostForm("bg")) == "" {
		
		
		if ctx.User.ID == 0 {
			resp403(c)
			return
		}
		// เช็คแปปว่ายศสูงพอมั้ย
		var perms int
		db.QueryRow("SELECT perms FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&perms)
		// ลบคำเชิญเก่าออก
		var clan int
		db.QueryRow("SELECT clan FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&clan)
		if clan == 0 {
			resp403(c)
			return
		}
		
		db.Exec("DELETE FROM clans_invites WHERE clan = ?", clan)
		
		var s string
		
		s = randSeq(8)

		db.Exec("INSERT INTO clans_invites(clan, invite) VALUES (?, ?)", clan, s)
	} else {
		// เช็คแปปว่ายศสูงพอมั้ย
		var perms int
		db.QueryRow("SELECT perms FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&perms)
		// ลบคำเชิญเก่าออก
		var clan int
		db.QueryRow("SELECT clan FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&clan)
		if clan == 0 {
			resp403(c)
			return
		}
		
		tag := "0"
		if c.PostForm("tag") != "" {
			tag = c.PostForm("tag")
		}
		
		if db.QueryRow("SELECT 1 FROM clans WHERE tag = ? AND id != ?", c.PostForm("tag"), clan).
		Scan(new(int)) != sql.ErrNoRows {
			resp403(c)
			addMessage(c, errorMessage{T(c, "เหมือนจะมีคนใช้ตัวย่อนั้นไปแล้วนะ...")})
			return
		}
		
		db.Exec("UPDATE clans SET description = ?, icon = ?, tag = ?, background = ? WHERE id = ?", c.PostForm("description"), c.PostForm("icon"), tag, c.PostForm("bg"), clan)
	}
	addMessage(c, successMessage{T(c, "เสร็จสิ้น!")})
	getSession(c).Save()
	c.Redirect(302, "/settings/clansettings")
}


func clanInvite(c *gin.Context) {
	i := c.Param("inv")
	
	res := resolveInvite(i)
	s := strconv.Itoa(res)
	if res != 0 {
	
		// ไอ้บ้านี้มันล็อกอินยัง
		if getContext(c).User.ID == 0 {
			resp403(c)
			return
		}
	
		// เห้ยไอ้นี้โดนแบนปะวะ
		if getContext(c).User.Privileges & 1 != 1 {
			resp403(c)
			return
		}
		
		// มีแคลนนี้ปะวะเนี่ย
			if db.QueryRow("SELECT 1 FROM clans WHERE id = ?", res).
			Scan(new(int)) == sql.ErrNoRows {

				addMessage(c, errorMessage{T(c, "ไม่เห็นจะมีแคลนนี้เลย!")})
				getSession(c).Save()
				c.Redirect(302, "/c/"+s)
				return
			}
		// ไอ้เหี้ยนี้อยู่ในแคลนปะวะ?
			if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ?", getContext(c).User.ID).
			Scan(new(int)) != sql.ErrNoRows {
				
				addMessage(c, errorMessage{T(c, "เหมือนว่าคุณน่าจะอยู่ในแคลนอยู่แล้วนะ")})
				getSession(c).Save()
				c.Redirect(302, "/c/"+s)
				return
			}

		// ควยไรสัส
		var count int
		var limit int
		// เช็คคน
		db.QueryRow("SELECT COUNT(*) FROM user_clans WHERE clan = ? ", res).Scan(&count)
		db.QueryRow("SELECT mlimit FROM clans WHERE id = ? ", res).Scan(&limit)
		if count >= limit {
			addMessage(c, errorMessage{T(c, "โทษทีนะ แคลนนี้เต็มแล้วอะ ;w;")})
			getSession(c).Save()
			c.Redirect(302, "/c/"+s)
			return
		}
		// เข้าแคลน
		db.Exec("INSERT INTO `user_clans`(user, clan, perms) VALUES (?, ?, 1);", getContext(c).User.ID, res)
		addMessage(c, successMessage{T(c, "เข้าแคลนแล้วจ้า!")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+s)
	} else {
		resp403(c)
		addMessage(c, errorMessage{T(c, "ไม่!!!")})
	}
}

func clanKick(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}

	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND perms = 8", getContext(c).User.ID).
			Scan(new(int)) == sql.ErrNoRows {
				resp403(c)
				return
			}

	member, err := strconv.ParseInt(c.PostForm("member"), 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	if member == 0 {
		resp403(c)
		return
	}

	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND perms = 1", member).
			Scan(new(int)) == sql.ErrNoRows {
				resp403(c)
				return
			}

	db.Exec("DELETE FROM user_clans WHERE user = ?", member)
	addMessage(c, successMessage{T(c, "เสร็จสิ้น!")})
	getSession(c).Save()
	c.Redirect(302, "/settings/clansettings")
}

func resolveInvite(c string)(int) {
	var clanid int
	row := db.QueryRow("SELECT clan FROM clans_invites WHERE invite = ?", c)
		err := row.Scan(&clanid)
		
		if err != nil {
			fmt.Println(err)
		}
	fmt.Println(clanid)
	return clanid
}
