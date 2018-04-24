package proxy

import (
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/parnurzeal/gorequest"
	"github.com/stakater/JenkinsProxy/pkg/parser"
	"github.com/stakater/JenkinsProxy/pkg/providers"
)

type Proxy struct {
	provider     string
	upstreamURL  string
	allowedPaths []string
	secret       string
}

func (p *Proxy) isPathAllowed(path string) bool {
	// All paths allowed
	if len(p.allowedPaths) == 0 {
		return true
	}

	// Check if given passed exists in allowedPaths
	for _, p := range p.allowedPaths {
		if strings.TrimSuffix(strings.TrimSpace(p), "/") ==
			strings.TrimSuffix(strings.TrimSpace(path), "/") {
			return true
		}
	}
	return false
}

func (p *Proxy) redirect(hook *providers.Hook, path string) (gorequest.Response, []error) {
	// Set SetDoNotClearSuperAgent to true so that request
	// agent is not reset on POST call
	request := gorequest.New().SetDoNotClearSuperAgent(true)

	// Set Headers from hook
	for key, value := range hook.Headers {
		request.AppendHeader(key, value)
	}

	resp, _, errs := request.Post(p.upstreamURL + path).Send(hook.Payload).End()

	return resp, errs
}

func (p *Proxy) proxyRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	if !p.isPathAllowed(r.URL.Path) {
		log.Printf("Not allowed to proxy path: '%s'", r.URL.Path)
		http.Error(w, "Not allowed to proxy path: '"+r.URL.Path+"'", http.StatusForbidden)
		return
	}

	provider, err := providers.NewProvider(p.provider, p.secret)
	if err != nil {
		log.Printf("Error creating provider: %s", err)
		http.Error(w, "Error creating Provider", http.StatusInternalServerError)
		return
	}

	hook, err := parser.Parse(r, provider)
	if err != nil {
		log.Printf("Eror Parsing Hook: %s", err)
		http.Error(w, "Error parsing Hook: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !provider.Validate(*hook) {
		log.Printf("Eror Validating Hook: %s", err)
		http.Error(w, "Error validating Hook", http.StatusBadRequest)
		return
	}

	resp, errs := p.redirect(hook, r.URL.Path)
	if errs != nil {
		log.Printf("Error Redirecting '%s' to upstream '%s': %s\n", r.URL.Path, p.upstreamURL, errs)
		http.Error(w, "Error Redirecting '"+r.URL.Path+"' to upstream '"+p.upstreamURL+"'", http.StatusInternalServerError)
		return
	}

	log.Printf("Redirected incomming request '%s' to '%s' with Response: '%s'\n",
		r.URL.Path, p.upstreamURL, resp.Status)
}

// Health Check Endpoint
func (p *Proxy) health(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.WriteHeader(200)
	w.Write([]byte("I'm Healthy and I know it! ;) "))
}

func NewProxy(listenAddress string, upstreamURL string, allowedPaths []string,
	provider string, secret string) {
	// Validate Params
	if len(strings.TrimSpace(listenAddress)) == 0 {
		panic("Cannot create Proxy with empty listenAddress")
	}
	if len(strings.TrimSpace(secret)) == 0 {
		panic("Cannot create Proxy with empty secret")
	}
	if len(strings.TrimSpace(upstreamURL)) == 0 {
		panic("Cannot create Proxy with empty upstreamURL")
	}

	proxy := Proxy{
		provider:     provider,
		upstreamURL:  upstreamURL,
		allowedPaths: allowedPaths,
		secret:       secret,
	}

	router := httprouter.New()
	router.GET("/health", proxy.health)
	router.POST("/*path", proxy.proxyRequest)

	log.Printf("Listening at: %s", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, router))
}
