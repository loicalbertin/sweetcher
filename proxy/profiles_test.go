package proxy

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func makeURL(t *testing.T, rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Error(err)
	}
	return u
}

func TestProfile_chooseProxy(t *testing.T) {
	p1 := makeURL(t, "http://myproxy.mycomp.it:8080")
	p2 := makeURL(t, "http://hiddenproxy.mycomp.it:8080")
	type fields struct {
		Default *url.URL
		Rules   []Rule
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *url.URL
		wantErr bool
	}{
		{"TestNoMatchWithDirectDefault",
			fields{nil, []Rule{Rule{Pattern: "*.google.com", Proxy: p1}}},
			args{&http.Request{URL: makeURL(t, "http://somewhere.else")}},
			nil, false},
		{"TestNoMatchWithDefaultProxy",
			fields{p1, []Rule{Rule{Pattern: "*.google.com", Proxy: nil}}},
			args{&http.Request{URL: makeURL(t, "http://somewhere.else")}},
			p1, false},
		{"TestMatchSubDomain",
			fields{nil, []Rule{Rule{Pattern: "*.google.com", Proxy: p1}}},
			args{&http.Request{URL: makeURL(t, "http://test.google.com")}},
			p1, false},
		{"TestMatchSubDomainWithPath",
			fields{nil, []Rule{Rule{Pattern: "*.google.com", Proxy: p1}}},
			args{&http.Request{URL: makeURL(t, "http://test.google.com/login/user")}},
			p1, false},
		{"TestMatchStar",
			fields{nil, []Rule{Rule{Pattern: "*google.com", Proxy: p1}}},
			args{&http.Request{URL: makeURL(t, "http://test.google.com")}},
			p1, false},
		{"TestMatchStar2",
			fields{nil, []Rule{Rule{Pattern: "*google.com", Proxy: p1}}},
			args{&http.Request{URL: makeURL(t, "http://testgoogle.com")}},
			p1, false},
		{"TestFirstMatchToProxy",
			fields{nil, []Rule{
				Rule{Pattern: "*.google.com", Proxy: p1},
				Rule{Pattern: "*.com", Proxy: p2},
			}},
			args{&http.Request{URL: makeURL(t, "http://test.google.com")}},
			p1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Default: tt.fields.Default,
				Rules:   tt.fields.Rules,
			}
			got, err := p.chooseProxy(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Profile.chooseProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Profile.chooseProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}
