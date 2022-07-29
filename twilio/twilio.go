package twilio

import (
	"log"

	twl "github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

type (
	Twilio struct {
		cln *twl.RestClient
		ssi string
	}
	Options struct {
		SID        string
		Token      string
		ServiceSID string
	}
)

func New(o *Options) *Twilio {
	if o.SID == "" {
		log.Fatal("twilio - SID option not provided")
	}
	if o.Token == "" {
		log.Fatal("twilio - Token option not provided")
	}
	if o.ServiceSID == "" {
		log.Fatal("twilio - ServiceSID option not provided")
	}

	c := twl.NewRestClientWithParams(twl.ClientParams{
		Username: o.SID,
		Password: o.Token,
	})

	t := &Twilio{c, o.ServiceSID}
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
