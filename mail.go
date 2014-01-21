package main

import "net/smtp"

   
func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	if (a != nil) {
		return smtp.SendMail(addr, a, from, to, msg)
	}
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Text.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}