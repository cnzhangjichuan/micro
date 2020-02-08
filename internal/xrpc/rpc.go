package xrpc

import (
	"github.com/cnzhangjichuan/micro/xutils"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, _, err := w.(http.Hijacker).Hijack()

	if err != nil {
		logError("rpc hijacking [%s] error: %v", r.RemoteAddr, err)
		return
	}
	defer conn.Close()

	// verification
	ac := env.aut.Create()
	packet := xutils.NewPacket(1024)
	packet.BeginWrite()
	packet.Write(ac)
	packet.EndWrite()
	packet.FlushToConn(conn)
	packet.ReadConn(conn)
	if !env.aut.Check(ac, packet.Data()) {
		packet.BeginWrite()
		packet.WriteByte(xPREFIX_ERR)
		packet.WriteString(`服务未授权`)
		packet.EndWrite()
		packet.FlushToConn(conn)
		return
	}

	packet.BeginWrite()
	packet.WriteByte(xPREFIX_DATA)
	packet.WriteString(`OK`)
	packet.EndWrite()
	packet.FlushToConn(conn)

	// process request
	dpo := &rpcDpo{}
	for err == nil {
		err = packet.ReadConn(conn)
		if err != nil {
			break
		}
		handler, ok := env.handlers[packet.ReadString()]
		if !ok {
			packet.BeginWrite()
			packet.WriteByte(xPREFIX_ERR)
			packet.WriteString(`服务不存在或已删除`)
			packet.EndWrite()
			_, err = packet.FlushToConn(conn)
			continue
		}
		dpo.data = packet.Data()
		err = handler.Func(dpo)
		if err != nil {
			packet.BeginWrite()
			packet.WriteByte(xPREFIX_ERR)
			packet.WriteString(err.Error())
			packet.EndWrite()
			_, err = packet.FlushToConn(conn)
			continue
		}

		resp, err := xutils.MarshalJson(dpo.resp)
		if err != nil {
			packet.BeginWrite()
			packet.WriteByte(xPREFIX_ERR)
			packet.WriteString(err.Error())
			packet.EndWrite()
			_, err = packet.FlushToConn(conn)
			continue
		}

		packet.BeginWrite()
		packet.WriteByte(xPREFIX_DATA)
		packet.Write(resp)
		packet.EndWrite()
		_, err = packet.FlushToConn(conn)
	}
}
