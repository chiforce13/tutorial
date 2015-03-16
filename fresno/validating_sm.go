package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	//it is customary to use the named import version for seven5 as "s5"
	"github.com/coocood/qbs"
	s5 "github.com/seven5/seven5"

	"tutorial/shared"
)

//passwordResetRequest is a struct that is useful to store in the database
//when you have sent a password reset request.  The UserRecord field is
//the primary key, and should be the UDID of the user record; since it is
//a primary key, there can be only one outstanding request at a time for a
//user. The Request is encoded as a udid in the Request field.  The MailLog
//is just a general purpose place to record information about what was
//emailed to the user, if anything.  The SentAt field is a time that
//the request was sent if you would like to expire the request.
type passwordResetRequest struct {
	UserRecord string `qbs:"pk"`
	MailLog    string
	Request    string
	SentAt     time.Time
}

// fresnoValidatingSessionManager knows how to check credentials, interact
// with sessions, and issue/reclaim password reset requests.
type fresnoValidatingSessionManager struct {
	*s5.SimpleSessionManager
}

// newFresnoValidatingSessionManager creates a new FresnoValidatingSessionManager
// but returns it as an s5.ValidatingSessionManager to insure that we meet the
// interface it requires.
func newFresnoValidatingSessionManager() s5.ValidatingSessionManager {
	result := &fresnoValidatingSessionManager{}
	result.SimpleSessionManager = s5.NewSimpleSessionManager(result)
	return result
}

//
// Generate converts a unique id, previosuly generated by this session manager
// into a user data record.
//
func (self *fresnoValidatingSessionManager) Generate(uniqId string) (interface{}, error) {
	q, err := qbs.GetQbs()
	if err != nil {
		return nil, err
	}
	var ur shared.UserRecord
	ur.UserUdid = uniqId
	if err := q.Find(&ur); err != nil {
		return nil, err
	}
	return &ur, nil
}

// Check that a username and password are as we have them in the database.
// If they match, we return the user's UDID as the uniq value for the session
// plus the user data record.
func (self *fresnoValidatingSessionManager) ValidateCredentials(username, pwd string) (string, interface{}, error) {
	q, err := qbs.GetQbs()
	if err != nil {
		return "", nil, err
	}
	defer q.Close()

	var ur shared.UserRecord
	u := strings.TrimSpace(username)
	p := strings.TrimSpace(pwd)
	if len(u) == 0 || len(p) == 0 {
		return "", nil, err
	}
	cond := qbs.NewEqualCondition("email_addr", u).AndEqual("password", p)
	if err := q.Condition(cond).Find(&ur); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error trying to validate credentials: %v", err)
			return "", nil, err
		}
		//normal case of bad pwd
		return "", nil, nil
	}
	//return the udid as uniq part,then the rest of the object as user data
	return ur.UserUdid, &ur, nil
}

// SendUserDetails is responsible for filtering out fields that we may not wish
// to send to the client side of the wire when returning a user record.
func (self *fresnoValidatingSessionManager) SendUserDetails(i interface{}, w http.ResponseWriter) error {
	ur := i.(*shared.UserRecord)
	ur.Password = ""
	return s5.SendJson(w, ur)
}

// GenerateResetRequest takes a user account (udid) and generates the side effect
// to inform that user that they have a one-time token for reseting their password.
// In a real application, it would probably send email but this version just prints
// the magic url on the console.
func (self *fresnoValidatingSessionManager) GenerateResetRequest(user string) (string, error) {
	q, err := qbs.GetQbs()
	if err != nil {
		return "", err
	}
	defer q.Close()

	resetToken := s5.UDID()
	var prr passwordResetRequest
	prr.UserRecord = user

	//maybe they have one pending?
	err = q.Find(&prr)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[AUTH] unable to read password reset request %v", err)
		return "", err
	}
	if err == nil {
		log.Printf("[AUTH] invalidating previous request %s at %v", prr.Request, prr.SentAt)
	}

	prr.Request = resetToken
	prr.SentAt = time.Now()
	prr.MailLog = "didn't send email"

	if _, err := q.Save(&prr); err != nil {
		log.Printf("[AUTH] unable to save password reset request %+v: %v", prr, err)
		return "", err
	}

	log.Printf("fresno doesn't send email, so paste this in your browser\n%s",
		fmt.Sprintf("%s/%s/%s/%s", PWD_RESET_PREFIX, user, resetToken, PWD_RESET_PAGE))
	return resetToken, nil
}

// Use reset request tries to confirm that a given user has previously been
// issued the provided request id.  This checks the database and, if the request
// is found, it deletes it so it cannot be used again. It does not expire
// requests by refusing attempts to use "old" requests.
func (self *fresnoValidatingSessionManager) UseResetRequest(user string, req string, newpwd string) (bool, error) {
	q, err := qbs.GetQbs()
	if err != nil {
		return false, err
	}
	defer q.Close()
	var prr passwordResetRequest
	prr.UserRecord = user

	err = q.Find(&prr)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[AUTH] unable to read password reset request %v", err)
		return false, err
	}
	if err == sql.ErrNoRows {
		log.Printf("[AUTH] no previous password reset found (%s tried)", req)
		return false, nil
	}
	if prr.Request != req {
		log.Printf("[AUTH] bad password reset found (%s != %s)", prr.Request, req)
		return false, nil
	}
	//if there is no new password, we are done
	if newpwd == "" {
		return true, nil
	}

	//reset the password to the provided value and then delete the request
	var ur shared.UserRecord
	ur.UserUdid = user
	err = q.Find(&ur)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == sql.ErrNoRows {
		log.Printf("[AUTH] bad user id provided (%s)", user)
		return false, nil
	}
	ur.Password = newpwd
	if _, err := q.Save(&ur); err != nil {
		log.Printf("[AUTH] could not save updated user (%s)", user)
		return false, err
	}
	if _, err := q.Delete(&prr); err != nil {
		log.Printf("[AUTH] could not delete password reset request (user %s) %s", user, req)
		return false, err
	}
	return true, nil
}
