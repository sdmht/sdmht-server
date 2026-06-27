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

func TestSubscriptionOnlineCount(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)
	count := 0
	for range 100 {
		go func() {
			sub := c.Websocket("subscription { onlineCount }")
			var msg map[string]int32
			for j := 0; j <= 100; j++ {
				a.NoError(sub.Next(&msg))
				a.EqualValues(j, msg["onlineCount"])
			}
			for j := 99; j >= 0; j-- {
				a.NoError(sub.Next(&msg))
				a.EqualValues(j, msg["onlineCount"])
			}
			a.NoError(sub.Close())
			count++
		}()
	}
	time.Sleep(100 * time.Millisecond)
	for i := range 100 {
		go func() {
			hb := c.Websocket("subscription heartbeat($uid: String!) { heartbeat(uid: $uid) }", client.Var("uid", i))
			time.Sleep(1000 * time.Millisecond)
			a.NoError(hb.Close())
		}()
		time.Sleep(9 * time.Millisecond)
	}
	time.Sleep(2000 * time.Millisecond)
	a.Equal(100, count)
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

// 测试缓存发现功能
// 第一部分：A先使用cachedResources保存['a.txt']，然后B调用cachedResourcePeers查a.txt，此时B那边应该返回包含A的uid的列表
func TestCachedResourceDiscovery(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)

	// A 使用 cachedResources 保存资源路径
	var mutationResult map[string]any
	err := c.Post(`mutation { cachedResources(uid: "A", paths: ["a.txt", "b.txt"]) }`, &mutationResult)
	a.NoError(err)

	// B 查询 a.txt 的对等节点
	var queryResult map[string]any
	err = c.Post(`query { cachedResourcePeers(uid: "B", path: "a.txt") }`, &queryResult)
	a.NoError(err)
	peers := queryResult["cachedResourcePeers"].([]any)
	a.Equal(1, len(peers))
	a.Equal("A", peers[0])

	// B 查询 b.txt 的对等节点
	err = c.Post(`query { cachedResourcePeers(uid: "B", path: "b.txt") }`, &queryResult)
	a.NoError(err)
	peers = queryResult["cachedResourcePeers"].([]any)
	a.Equal(1, len(peers))
	a.Equal("A", peers[0])

	// B 查询不存在的资源
	err = c.Post(`query { cachedResourcePeers(uid: "B", path: "c.txt") }`, &queryResult)
	a.NoError(err)
	peers = queryResult["cachedResourcePeers"].([]any)
	a.Equal(0, len(peers))

	// A 查询 a.txt 时应该排除自己
	err = c.Post(`query { cachedResourcePeers(uid: "A", path: "a.txt") }`, &queryResult)
	a.NoError(err)
	peers = queryResult["cachedResourcePeers"].([]any)
	a.Equal(0, len(peers))

	// C 也保存 a.txt
	err = c.Post(`mutation { cachedResources(uid: "C", paths: ["a.txt"]) }`, &mutationResult)
	a.NoError(err)

	// B 再次查询 a.txt，应该返回 A 和 C
	err = c.Post(`query { cachedResourcePeers(uid: "B", path: "a.txt") }`, &queryResult)
	a.NoError(err)
	peers = queryResult["cachedResourcePeers"].([]any)
	a.Equal(2, len(peers))
	peerSet := make(map[string]bool)
	for _, p := range peers {
		peerSet[p.(string)] = true
	}
	a.True(peerSet["A"])
	a.True(peerSet["C"])
}

// 测试信令连接功能
// 第二部分：A和B调用signaling保持监听，B使用sendSignaling向A发送连接数据并告知自己的uid请求连接，
// 此时A的signaling应该返回数据且连接不关闭继续监听，A也通过sendSignaling向B发送数据，再互相重复一次发送数据接收数据，连接成功建立
func TestSignalingConnection(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)

	// A 和 B 调用 listenSignaling 保持监听
	sigA := c.Websocket(`subscription { listenSignaling(uid: "A") }`)
	sigB := c.Websocket(`subscription { listenSignaling(uid: "B") }`)

	// B 使用 sendSignaling 向 A 发送连接数据
	var sendResult map[string]any
	bConnectData := map[string]any{"type": "connect", "message": "hello from B"}
	sendQuery := `subscription sendSignaling($uid: String!, $to: String!, $data: JSON!) { sendSignaling(uid: $uid, to: $to, data: $data) }`

	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "B"),
		client.Var("to", "A"),
		client.Var("data", bConnectData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// A 的 listenSignaling 应该收到数据
	var msgA map[string]any
	a.NoError(sigA.Next(&msgA))
	signalingData := msgA["listenSignaling"].(map[string]any)
	a.Equal("B", signalingData["uid"])
	a.Equal(bConnectData, signalingData["data"])

	// A 也通过 sendSignaling 向 B 发送数据
	aConnectData := map[string]any{"type": "ack", "message": "hello from A"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "A"),
		client.Var("to", "B"),
		client.Var("data", aConnectData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// B 的 listenSignaling 应该收到数据
	var msgB map[string]any
	a.NoError(sigB.Next(&msgB))
	signalingData = msgB["listenSignaling"].(map[string]any)
	a.Equal("A", signalingData["uid"])
	a.Equal(aConnectData, signalingData["data"])

	// 再互相重复一次发送数据接收数据
	// B 再次发送数据
	bSecondData := map[string]any{"type": "confirm", "message": "connection established"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "B"),
		client.Var("to", "A"),
		client.Var("data", bSecondData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// A 再次收到数据
	a.NoError(sigA.Next(&msgA))
	signalingData = msgA["listenSignaling"].(map[string]any)
	a.Equal("B", signalingData["uid"])
	a.Equal(bSecondData, signalingData["data"])

	// A 再次发送数据
	aSecondData := map[string]any{"type": "confirm", "message": "connection confirmed"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "A"),
		client.Var("to", "B"),
		client.Var("data", aSecondData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// B 再次收到数据
	a.NoError(sigB.Next(&msgB))
	signalingData = msgB["listenSignaling"].(map[string]any)
	a.Equal("A", signalingData["uid"])
	a.Equal(aSecondData, signalingData["data"])

	// 关闭 listenSignaling 订阅
	a.NoError(sigA.Close())
	a.NoError(sigB.Close())
}

// 测试完整的缓存资源发现和信令连接流程
func TestCachedResourceAndSignalingIntegration(t *testing.T) {
	a := assert.New(t)
	srv := setupGraphqlService()
	c := client.New(srv)

	// 第一部分：发现
	// A 先使用 cachedResources 保存资源
	var mutationResult map[string]any
	err := c.Post(`mutation { cachedResources(uid: "playerA", paths: ["resource.txt"]) }`, &mutationResult)
	a.NoError(err)

	// B 调用 cachedResourcePeers 查询 resource.txt
	var queryResult map[string]any
	err = c.Post(`query { cachedResourcePeers(uid: "playerB", path: "resource.txt") }`, &queryResult)
	a.NoError(err)
	peers := queryResult["cachedResourcePeers"].([]any)
	a.Equal(1, len(peers))
	targetUID := peers[0].(string)
	a.Equal("playerA", targetUID)

	// 第二部分：连接
	// A 和 B 调用 listenSignaling 保持监听
	sigA := c.Websocket(`subscription { listenSignaling(uid: "playerA") }`)
	sigB := c.Websocket(`subscription { listenSignaling(uid: "playerB") }`)

	// B 使用 sendSignaling 向 A 发送连接数据
	var sendResult map[string]any
	connectData := map[string]any{"type": "offer", "sdp": "offer_sdp_data"}
	sendQuery := `subscription sendSignaling($uid: String!, $to: String!, $data: JSON!) { sendSignaling(uid: $uid, to: $to, data: $data) }`

	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "playerB"),
		client.Var("to", targetUID),
		client.Var("data", connectData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// A 收到 B 的连接请求
	var msgA map[string]any
	a.NoError(sigA.Next(&msgA))
	signalingData := msgA["listenSignaling"].(map[string]any)
	a.Equal("playerB", signalingData["uid"])
	a.Equal(connectData, signalingData["data"])

	// A 回复连接确认
	answerData := map[string]any{"type": "answer", "sdp": "answer_sdp_data"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "playerA"),
		client.Var("to", "playerB"),
		client.Var("data", answerData),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// B 收到 A 的回复
	var msgB map[string]any
	a.NoError(sigB.Next(&msgB))
	signalingData = msgB["listenSignaling"].(map[string]any)
	a.Equal("playerA", signalingData["uid"])
	a.Equal(answerData, signalingData["data"])

	// 第三部分：C向A发起连接（C没有缓存资源）
	// C 查询发现 A
	err = c.Post(`query { cachedResourcePeers(uid: "playerC", path: "resource.txt") }`, &queryResult)
	a.NoError(err)
	peers = queryResult["cachedResourcePeers"].([]any)
	a.Equal(1, len(peers)) // 只有 A 缓存了资源
	a.Equal("playerA", peers[0].(string))

	// C 创建 listenSignaling 订阅
	sigC := c.Websocket(`subscription { listenSignaling(uid: "playerC") }`)

	// C 向 A 发送连接请求
	offerDataC := map[string]any{"type": "offer", "sdp": "offer_sdp_from_c"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "playerC"),
		client.Var("to", "playerA"),
		client.Var("data", offerDataC),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// A 收到 C 的连接请求
	a.NoError(sigA.Next(&msgA))
	signalingData = msgA["listenSignaling"].(map[string]any)
	a.Equal("playerC", signalingData["uid"])
	a.Equal(offerDataC, signalingData["data"])

	// A 回复连接确认
	answerDataC := map[string]any{"type": "answer", "sdp": "answer_sdp_for_c"}
	c.WebsocketOnce(sendQuery, &sendResult,
		client.Var("uid", "playerA"),
		client.Var("to", "playerC"),
		client.Var("data", answerDataC),
	)
	a.Equal(nil, sendResult["sendSignaling"])

	// C 收到 A 的回复
	var msgC map[string]any
	a.NoError(sigC.Next(&msgC))
	signalingDataC := msgC["listenSignaling"].(map[string]any)
	a.Equal("playerA", signalingDataC["uid"])
	a.Equal(answerDataC, signalingDataC["data"])

	// 关闭 listenSignaling 订阅
	a.NoError(sigA.Close())
	a.NoError(sigB.Close())
	a.NoError(sigC.Close())
}
