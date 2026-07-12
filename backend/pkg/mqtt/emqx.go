package mqtt

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// topicHandler 保存主题和对应的处理器
type topicHandler struct {
	Topic   string
	Handler func(topic string, payload []byte)
}

// EMQXClient 封装了 EMQX 的发布订阅操作，支持重连后自动重新订阅
type EMQXClient struct {
	client       mqtt.Client
	broker       string
	subscriptions []topicHandler
	mu           sync.RWMutex
}

// NewEMQXClient 新建一个 EMQX 客户端
func NewEMQXClient(broker, username, password, clientID string) (*EMQXClient, error) {
	if clientID == "" {
		clientID = "eiot-backend-" + time.Now().Format("20060102150405")
	}
	e := &EMQXClient{broker: broker}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(60 * time.Second).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("[EMQX] connection lost: %v", err)
		}).
		// 重连成功后自动重新订阅所有主题
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("[EMQX] connected, re-subscribing %d topics", len(e.subscriptions))
			e.mu.RLock()
			subs := make([]topicHandler, len(e.subscriptions))
			copy(subs, e.subscriptions)
			e.mu.RUnlock()
			for _, sub := range subs {
				e.doSubscribe(sub.Topic, sub.Handler)
			}
		})

	cli := mqtt.NewClient(opts)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	log.Printf("[EMQX] connected to %s", broker)
	e.client = cli
	return e, nil
}

// doSubscribe 实际执行订阅
func (e *EMQXClient) doSubscribe(topic string, handler func(topic string, payload []byte)) error {
	token := e.client.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		handler(m.Topic(), m.Payload())
	})
	token.Wait()
	if err := token.Error(); err != nil {
		log.Printf("[EMQX] subscribe %s failed: %v", topic, err)
		return err
	}
	log.Printf("[EMQX] subscribed: %s", topic)
	return nil
}

// Subscribe 订阅一个主题，支持主题通配符（+、#）
func (e *EMQXClient) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	if e == nil || e.client == nil {
		return nil
	}
	// 保存到列表，以便重连后重新订阅
	e.mu.Lock()
	e.subscriptions = append(e.subscriptions, topicHandler{Topic: topic, Handler: handler})
	e.mu.Unlock()
	return e.doSubscribe(topic, handler)
}

// Publish 发布一条消息（JSON 格式）
func (e *EMQXClient) Publish(topic string, payload interface{}) error {
	if e == nil || e.client == nil {
		return nil
	}
	data, _ := json.Marshal(payload)
	token := e.client.Publish(topic, 1, false, data)
	token.Wait()
	if err := token.Error(); err != nil {
		log.Printf("[EMQX] publish %s failed: %v", topic, err)
		return err
	}
	return nil
}

// IsConnected 检查连接状态
func (e *EMQXClient) IsConnected() bool {
	return e != nil && e.client != nil && e.client.IsConnected()
}
