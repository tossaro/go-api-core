package twilio

import (
	twl "github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

type Twilio struct {
	cln *twl.RestClient
	ssi string
}

func New(s string, tkn string, sid string) *Twilio {
	c := twl.NewRestClientWithParams(twl.ClientParams{
		Username: s,
		Password: tkn,
	})

	t := &Twilio{c, sid}
	return t
}

func (t *Twilio) Verify(p string) (*string, error) {
	params := &verify.CreateVerificationParams{}
	params.SetChannel("sms")
	params.SetTo(p)

	v, err := t.cln.VerifyV2.CreateVerification(t.ssi, params)
	if err != nil {
		return nil, err
	}
	return v.Sid, nil
}

func (t *Twilio) VerifyCheck(p string, c string) (*string, error) {
	params := &verify.CreateVerificationCheckParams{}
	params.SetCode(c)
	params.SetTo(p)

	v, err := t.cln.VerifyV2.CreateVerificationCheck(t.ssi, params)
	if err != nil {
		return nil, err
	}
	return v.Sid, nil
}
