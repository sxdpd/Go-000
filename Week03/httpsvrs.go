package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	serverNum = 3
	beginPort = 8080
)

type ShutDownHandler struct {
	closeCh chan string
}

func (h *ShutDownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "All servers are shutting down")

	select {
	case <-h.closeCh:
	default:
		h.closeCh <- fmt.Sprintf("shutting down from server %s%s", r.Host, r.URL.Path)
	}
}

func newHttpServer(port int, closeCh chan string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, this is server %s", r.Host)
	}))
	mux.Handle("/shutdown", &ShutDownHandler{closeCh})

	return &http.Server{Handler: mux,
		Addr: fmt.Sprintf(":%d", port)}
}

func main() {
	g, _ := errgroup.WithContext(context.Background())

	// 监听系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

	// 创建 http 服务关闭通道
	serverCloseCh := make(chan string, 1)

	// 创建 http 服务对象列表
	httpServers := make([]*http.Server, serverNum)
	for i := 0; i < serverNum; i++ {
		httpServers[i] = newHttpServer(beginPort+i, serverCloseCh)
	}

	// 启动多个 http 服务
	for i := 0; i < serverNum; i++ {
		i := i
		g.Go(func() error {
			server := httpServers[i]
			err := server.ListenAndServe()
			select {
			case <-serverCloseCh:
			default:
				serverCloseCh <- fmt.Sprintf("server %d err: %s", beginPort+i, err)
			}

			return nil
		})
	}

	// 监听所有关闭信号
	g.Go(func() error {
		select {
		case s := <-sigCh:
			log.Println("Received system signal:", s)
		case msg := <-serverCloseCh:
			log.Println("Received server close signal:", msg)
		}

		// 关闭所有通道
		signal.Stop(sigCh)
		close(serverCloseCh)

		// 关闭所有 http 服务
		for i := 0; i < serverNum; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			_ = httpServers[i].Shutdown(ctx)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Println("err:", err)
	} else {
		log.Println("Successfully shut down.")
	}
}
