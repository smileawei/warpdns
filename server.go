package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/miekg/dns"
)

type Server struct {
	cfg *Config
	doh *DoHClient
}

func NewServer(cfg *Config) (*Server, error) {
	doh, err := NewDoHClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, doh: doh}, nil
}

func (s *Server) Run(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.cfg.Listen)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	log.Printf("warpdns listening on udp %s -> %s%s (method=%s ecs=%v)",
		s.cfg.Listen, s.cfg.Upstream.URL, s.cfg.Upstream.Path,
		s.cfg.Upstream.Method, s.cfg.ECS.Enabled)

	var wg sync.WaitGroup
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				wg.Wait()
				return nil
			}
			log.Printf("read udp: %v", err)
			continue
		}
		pkt := make([]byte, n)
		copy(pkt, buf[:n])
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handle(ctx, conn, client, pkt)
		}()
	}
}

func (s *Server) handle(ctx context.Context, conn *net.UDPConn, client *net.UDPAddr, query []byte) {
	req := new(dns.Msg)
	if err := req.Unpack(query); err != nil {
		log.Printf("unpack from %s: %v", client, err)
		return
	}

	if s.cfg.ECS.Enabled {
		injectECS(req, s.cfg.ECS.Subnet)
	}

	packed, err := req.Pack()
	if err != nil {
		log.Printf("pack: %v", err)
		s.replyServFail(conn, client, req)
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, s.cfg.Upstream.Timeout.Std())
	defer cancel()

	respBytes, err := s.doh.Query(reqCtx, packed)
	if err != nil {
		log.Printf("doh query (qname=%s): %v", qname(req), err)
		s.replyServFail(conn, client, req)
		return
	}

	resp := new(dns.Msg)
	if err := resp.Unpack(respBytes); err != nil {
		log.Printf("unpack upstream response: %v", err)
		s.replyServFail(conn, client, req)
		return
	}
	// Preserve the original transaction ID in case upstream rewrote it.
	resp.Id = req.Id

	out, err := resp.Pack()
	if err != nil {
		log.Printf("pack response: %v", err)
		s.replyServFail(conn, client, req)
		return
	}
	if _, err := conn.WriteToUDP(out, client); err != nil {
		log.Printf("write udp to %s: %v", client, err)
	}
}

func (s *Server) replyServFail(conn *net.UDPConn, client *net.UDPAddr, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(req, dns.RcodeServerFailure)
	out, err := m.Pack()
	if err != nil {
		return
	}
	_, _ = conn.WriteToUDP(out, client)
}

func qname(m *dns.Msg) string {
	if len(m.Question) == 0 {
		return "<empty>"
	}
	return m.Question[0].Name
}
