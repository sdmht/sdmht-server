package main

import (
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/assert"
)

func TestQueryTime(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)
	var msg map[string]string
	now_time := time.Now()
	a.NoError(c.Post("{ time }", &msg))
	a.NotEmpty(msg["time"])
	// 解析时间
	parsed_time, err := time.Parse(time.RFC3339Nano, msg["time"])
	a.NoError(err)
	// 检测是否为当前时间，误差在10毫秒
	a.WithinDuration(parsed_time, now_time, time.Second+10*time.Millisecond)
}

func TestSubscriptionTime(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)
	sub := c.Websocket("subscription { time }")

	// 返回的是键为time，值为时间字符串的map
	var msg map[string]string
	a.NoError(sub.Next(&msg))
	a.NotEmpty(msg["time"])
	// 解析时间
	firstTime, err := time.Parse(time.RFC3339Nano, msg["time"])
	a.NoError(err)
	a.NoError(sub.Next(&msg))
	a.NotEmpty(msg["time"])
	secondTime, err := time.Parse(time.RFC3339Nano, msg["time"])
	a.NoError(err)
	// 检查时间是否在1秒内，误差在10毫秒
	a.WithinDuration(firstTime, secondTime, time.Second+10*time.Millisecond)
	a.NoError(sub.Close())
}

func TestSubscriptionOnline(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)
	oc := c.Websocket("subscription { onlineCount }")

	var msg map[string]any
	a.NoError(oc.Next(&msg))
	a.EqualValues(0, msg["onlineCount"])
	hb1 := c.Websocket("subscription { heartbeat(uid: \"1\") }")
	a.NoError(oc.Next(&msg))
	a.EqualValues(1, msg["onlineCount"])

	a.NoError(c.WebsocketOnce("subscription { listenAlive(uid: \"2\") }", &msg))
	a.Equal(nil, msg["listenAlive"])

	hb2 := c.Websocket("subscription { heartbeat(uid: \"2\") }")

	la2 := c.Websocket("subscription { listenAlive(uid: \"2\") }")

	la2_is_alive := true
	go func() {
		a.NoError(la2.Next(&msg))
		a.Equal(nil, msg["listenAlive"])
		a.NoError(la2.Close())
		la2_is_alive = false
	}()

	a.NoError(oc.Next(&msg))
	a.EqualValues(2, msg["onlineCount"])

	a.Equal(true, la2_is_alive)
	a.NoError(hb2.Close())
	time.Sleep(10 * time.Millisecond)
	a.Equal(false, la2_is_alive)

	a.NoError(oc.Next(&msg))
	a.EqualValues(1, msg["onlineCount"])
	a.NoError(hb1.Close())
	a.NoError(oc.Next(&msg))
	a.EqualValues(0, msg["onlineCount"])
	a.NoError(oc.Close())
}

func TestSubscriptionMatch(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)
	p1 := c.Websocket("subscription { matchOpponent(uid: \"1\", size: 5, version: \"0.0.1\")  }")
	p2 := c.Websocket("subscription { matchOpponent(uid: \"2\", size: 5, version: \"0.0.0\")  }")
	p3 := c.Websocket("subscription { matchOpponent(uid: \"3\", size: 4, version: \"0.0.1\")  }")
	p4 := c.Websocket("subscription { matchOpponent(uid: \"4\", size: 4, version: \"0.0.1\")  }")
	p5 := c.Websocket("subscription { matchOpponent(uid: \"5\", size: 4, version: \"0.0.1\")  }")
	// p1和p2与其他玩家因为参数不一致，无法匹配，p5因为没有对手也无法匹配，应当只有p3和p4可以匹配
	var msg_p1 map[string]any
	var msg_p2 map[string]any
	var msg_p3 map[string]any
	var msg_p4 map[string]any
	var msg_p5 map[string]any
	p1_is_matched := false
	p2_is_matched := false
	p3_is_matched := false
	p4_is_matched := false
	p5_is_matched := false
	go func() {
		a.Error(p1.Next(&msg_p1))
		p1_is_matched = true
	}()
	go func() {
		a.Error(p2.Next(&msg_p2))
		p2_is_matched = true
	}()
	go func() {
		a.NoError(p3.Next(&msg_p3))
		p3_is_matched = true
		a.Equal("4", msg_p3["matchOpponent"])
	}()
	go func() {
		a.NoError(p4.Next(&msg_p4))
		p4_is_matched = true
		a.Equal("", msg_p4["matchOpponent"])
		a.NoError(p4.Next(&msg_p4))
		a.Equal("3", msg_p4["matchOpponent"])
	}()
	go func() {
		a.Error(p5.Next(&msg_p5))
		p5_is_matched = true
	}()
	time.Sleep(10 * time.Millisecond)
	a.Equal(false, p1_is_matched)
	a.Equal(false, p2_is_matched)
	a.Equal(true, p3_is_matched)
	a.Equal(true, p4_is_matched)
	a.Equal(false, p5_is_matched)
	a.NoError(p1.Close())
	a.NoError(p2.Close())
	a.NoError(p5.Close())
	send_data_query := "subscription sendData($to:String!, $data: JSON!) { sendData(to: $to, data: $data) }"
	var msg_sd1 map[string]any
	c.WebsocketOnce(send_data_query, msg_sd1, client.Var("to", "3"), client.Var("data", map[string]any{"hello": "world"}))
	a.Equal(nil, msg_sd1["sendData"])
	a.NoError(p3.Next(&msg_p3))
	a.Equal(map[string]any{"hello": "world"}, msg_p3["matchOpponent"])
	var msg_sd2 map[string]any
	c.WebsocketOnce(send_data_query, &msg_sd2, client.Var("to", "4"), client.Var("data", map[string]any{"world": "hello"}))
	a.Equal(nil, msg_sd2["sendData"])
	a.NoError(p4.Next(&msg_p4))
	a.Equal(map[string]any{"world": "hello"}, msg_p4["matchOpponent"])
}
