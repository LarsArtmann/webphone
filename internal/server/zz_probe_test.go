package server

import (
    "context"
    "net/http/httptest"
    "testing"

    "github.com/larsartmann/webphone/internal/pbx"
)

func TestProbeFakeMailboxVerify(t *testing.T) {
    mailbox := &fakeMailbox{}
    srv := httptest.NewServer(mailbox.handler())
    t.Cleanup(srv.Close)
    client, err := pbx.NewClient(srv.URL)
    if err != nil { t.Fatal(err) }
    err = client.VerifyCredentials(context.Background(), pbx.Credentials{Extension: "1001", Password: "pw"})
    t.Logf("verify err: %v", err)
    if err != nil { t.Fatal(err) }
}
