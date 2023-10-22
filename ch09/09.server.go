package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/awoodbeck/gnp/ch09/handlers"
	"github.com/awoodbeck/gnp/ch09/middleware"
)

var (
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	cert  = flag.String("cert", "", "certificate")
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	flag.Parse()

	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}

func run(addr, files, cert, pkey string) error {
	// MUX 선언
	mux := http.NewServeMux()
	// /static/으로 시작하는 경우 주소에서 /static/을 떼어내고 .이 포함되지 않을 때만 라우팅
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(
				".", http.FileServer(http.Dir("./files")),
			),
		),
	)
	// 주소가 /로 시작하는 경우
	mux.Handle("/",
		handlers.Methods{
			// GET요청이 오는경우
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// http.ResponseWriter w가
					// http.Pusher 인터페이스를 구현하고 있는지 확인
					if pusher, ok := w.(http.Pusher); ok {
						targets := []string{
							"/static/style.css",
							"/static/hiking.svg",
						}
						for _, target := range targets {
							// 요청이 없어도 pusher로 클라이언트로 리소스 데이터 보내기
							if err := pusher.Push(target, nil); err != nil {
								// 실패시 테스트 콘솔 뿌리기
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}
					// "./files/index.html"을 찾아 클라이언트에 전송
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)

	// 서버 객체 선언
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	// 고루틴 생성
	go func() {
		// 시그널을 주고받는 c 채널 생성
		c := make(chan os.Signal, 1)
		// c 채널에 인터럽트 시그널 보내기
		signal.Notify(c, os.Interrupt)

		for {
			// 인터럽트를 받았다면
			if <-c == os.Interrupt {
				// 서버 닫기
				_ = srv.Close()
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	// 인증서랑 개인키가 있다면
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		// TLS로 서버 생성
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		// 아니면 일반 http로 서버 생성
		err = srv.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}
