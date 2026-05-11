// Command api 用户端 API 服务（监听 :17180）。
//
// 路由前缀：/api/v1
// 详见 docs/04-API规范.md §2。
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BFanSYe/image2api/backend/internal/bootstrap"
	"github.com/BFanSYe/image2api/backend/internal/router"
	"github.com/BFanSYe/image2api/backend/pkg/logger"
)

const serviceName = "api"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck(os.Args[2:]))
	}

	deps, err := bootstrap.Init(serviceName)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	r := router.New(router.Options{ServiceName: serviceName, Deps: deps})
	router.MountAPI(r, deps)

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(deps.Cfg.Server.APIPort),
		Handler:      r,
		ReadTimeout:  deps.Cfg.Server.ReadTimeout,
		WriteTimeout: deps.Cfg.Server.WriteTimeout,
	}

	if err := bootstrap.Run(srv, deps.Cfg.Server.ShutdownTimeout); err != nil {
		fmt.Println("server exit error:", err)
	}
}

func runHealthcheck(args []string) int {
	fs := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	addr := fs.String("addr", "127.0.0.1:17180", "HTTP listen address")
	path := fs.String("path", "/readyz", "health endpoint path")
	timeout := fs.Duration("timeout", 3*time.Second, "request timeout")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	url := strings.TrimSpace(*addr)
	if url == "" {
		fmt.Fprintln(os.Stderr, "empty healthcheck addr")
		return 2
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	endpoint := strings.TrimRight(url, "/") + "/" + strings.TrimLeft(*path, "/")

	client := &http.Client{Timeout: *timeout}
	resp, err := client.Get(endpoint)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "healthcheck returned HTTP %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
