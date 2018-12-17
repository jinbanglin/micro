package opts

import (
  "bytes"
  "context"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/go-redis/redis"
  "github.com/jinbanglin/go-micro"
  "github.com/jinbanglin/go-micro/client"
  "github.com/jinbanglin/go-micro/metadata"
  "github.com/jinbanglin/go-micro/registry"
  "github.com/jinbanglin/go-micro/server"
  "github.com/jinbanglin/go-plugins/codec/protorpc"
  "github.com/jinbanglin/go-plugins/transport/tcp"
  "github.com/jinbanglin/go-web"
  "github.com/jinbanglin/helper"
  "github.com/jinbanglin/log"
  "github.com/spf13/viper"
  "errors"
)

//client opts
func WClientOptions() (opts []client.Option) {
  return []client.Option{
    client.Transport(tcp.NewTransport()),
    client.Registry(registry.NewRegistry(registry.Addrs(viper.GetStringSlice("etcdv3.endpoints")...))),
    client.Codec("application/proto-rpc", protorpc.NewCodec),
    client.Retries(3),
    client.RequestTimeout(time.Second * 10),
  }
}

//socket server opts
func SServerOptions(name string) (opts []micro.Option) {
  return []micro.Option{
    micro.Name(name),
    micro.Registry(registry.NewRegistry(
      registry.Addrs(viper.GetStringSlice("etcdv3.endpoints")...),
    )),
    micro.RegisterTTL(time.Second * 30), //ttl 值必须大于 interval
    micro.RegisterInterval(time.Second * 15),
    micro.Transport(tcp.NewTransport()),
    micro.WrapHandler(HUB),
    micro.Version(viper.GetString("server.version")),
    micro.Metadata(map[string]string{"ip": helper.GetLocalIP()}),
  }
}

func SServerMakeClient(service micro.Service) client.Client {
  c := service.Client()
  c.Init(client.Retries(3), client.RequestTimeout(time.Second*10))
  return c
}

//web socket or http opts
func WServerWithOptions(name string, f func()) (opts []web.Option) {
  f()
  return []web.Option{
    web.Name(name),
    web.Registry(registry.NewRegistry(
      registry.Addrs(viper.GetStringSlice("etcdv3.endpoints")...),
    )),
    web.RegisterTTL(time.Second * 30),
    web.RegisterInterval(time.Second * 15),
  }
  return
}

type ResponseLogWriter struct {
  gin.ResponseWriter
  response *bytes.Buffer
}

func (w *ResponseLogWriter) Write(b []byte) (int, error) {
  w.response.Write(b)
  return w.ResponseWriter.Write(b)
}

var _G_PLACEHOLDER = []byte("l")

var LoadingError = errors.New("loading")
var ServerInternalError = errors.New("server internal error")

const REDIS_KEY_IDEMPOTENCY = "micro:idempotency:"

func HUB(next server.HandlerFunc) server.HandlerFunc {
  return func(ctx context.Context, req server.Request, rsp interface{}) error {
    now := time.Now()
    meta, _ := metadata.FromContext(ctx)
    soleID := meta["X-Sole-Id"]
    if helper.IsNilString(soleID) {
      return ServerInternalError
    }
    log.Infof2(ctx, "HUB_REQ |service=%s |method=%s "+
      "|address=%s |version=%s |content_type=%s |request=%v",
      req.Service()+"-"+server.DefaultServer.Options().Id,
      req.Method(),
      server.DefaultServer.Options().Metadata["ip"]+server.DefaultServer.Options().Address,
      server.DefaultServer.Options().Version,
      req.ContentType(),
      helper.Marshal2String(req.Request()),
    )
    b, err := helper.GRedisRing.GetSet(
      REDIS_KEY_IDEMPOTENCY+soleID,
      _G_PLACEHOLDER,
    ).Bytes()
    if err == nil && bytes.EqualFold(b, _G_PLACEHOLDER) {
      return LoadingError
    } else if err == nil {
      rsp = helper.Byte2String(b)
      return nil
    } else if err == redis.Nil {
      err = next(ctx, req, rsp)
      log.Infof2(ctx, "HUB_RSP |duration=%v |service=%s |method=%s "+
        "|address=%s |version=%s |content_type=%s |response=%v |err=%v",
        time.Since(now),
        req.Service()+"-"+server.DefaultServer.Options().Id,
        req.Method(),
        server.DefaultServer.Options().Metadata["ip"]+server.DefaultServer.Options().Address,
        server.DefaultServer.Options().Version,
        req.ContentType(),
        helper.Marshal2String(rsp),
        err,
      )
      //更新数据
      helper.GRedisRing.Set(
        REDIS_KEY_IDEMPOTENCY+soleID,
        helper.Marshal2Bytes(rsp),
        time.Second*20,
      )
    }
    return err
  }
}
