package deviate_dns

import (
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const RedirectTitle = "Deviate DNS"
const TxtRecordPrefix = "v=deviate-dns1 "
const TxtRecordKeyGoto = "goto"
const TxtRecordKeyEmail = "email"
const TxtRecordKeyKeepPath = "keeppath"
const TxtRecordKeyStatusCode = "statuscode"

func init() {
	caddy.RegisterModule(Middleware{})
}

// Middleware implements an HTTP handler that grabs redirect data from a txt record
type Middleware struct {
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.deviate-dns",
		New: func() caddy.Module { return new(Middleware) },
	}
}

func (g *Middleware) Provision(ctx caddy.Context) error {
	g.logger = ctx.Logger(g) // g.logger is a *zap.Logger
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	deviateRecord, err := getRecord(r.Host)
	w.Header().Set("Server", RedirectTitle)
	if err != nil {
		m.logger.Info(err.Error())
		w.Write([]byte("The domain is pointing to Deviate DNS. Error:" + err.Error()))
		return nil
	}
	if deviateRecord != nil {
		var location = "https://" + deviateRecord.Goto
		if deviateRecord.KeepPath {
			location = location + r.URL.Path
		}
		w.Header().Add("Location", location)
		w.WriteHeader(deviateRecord.StatusCode)
		w.Write([]byte("<center><h1>301 Moved Permanently</h1></center>\n<hr><center>" + RedirectTitle + "</center>"))
		return nil
	}
	return next.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)

// End of Caddy Module Setup

type TxtRecord struct {
	Goto string
	Email string
	StatusCode int
	KeepPath bool
}

func getRecord(name string) (*TxtRecord, error) {
	txtRecords, err := net.LookupTXT(name)
	if err != nil {
		return nil, err
	}
	var record = ""
	for _, txtRecord := range txtRecords {
		if strings.HasPrefix(txtRecord, TxtRecordPrefix) {
			txtRecord = strings.Replace(txtRecord, TxtRecordPrefix, "", -1)

			record = txtRecord
		}
	}
	if record == "" {
		err = errors.New("txt records found but none prefixed with " + TxtRecordPrefix)
		return nil, err
	}
	return parseRecord(record)
}

func parseRecord(record string) (*TxtRecord, error) {
	d := TxtRecord{StatusCode: 301, KeepPath: true}
	s := strings.Split(record, " ")
	for _, e := range s {
		parts := strings.Split(e, ":")
		switch parts[0] {
		case TxtRecordKeyGoto:
			d.Goto = parts[1]
		case TxtRecordKeyEmail:
			d.Email = parts[1]
		case TxtRecordKeyStatusCode:
			if i, err := strconv.Atoi(parts[1]); err == nil {
				d.StatusCode = i
			}
		case TxtRecordKeyKeepPath:
			d.KeepPath = parts[1] != "false"
		}
	}

	if d.Goto == "" || d.Email == "" {
		return nil, errors.New("deviate dns txt record found but missing goto or email value")
	}

	return &d, nil
}