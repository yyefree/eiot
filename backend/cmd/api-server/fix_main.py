import re

with open('main.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = """})

	log.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /sys/+/+/thing/event/property/batch_post, /sys/+/+/thing/event/property/history_post, /ota/device/inform/+/+, /sys/+/+/event/post, /sys/+/+/service/+/reply, /sys/+/+/rrpc/request/+")
}"""

new = """})

	// 飞燕标准 Topic：设备动态注册 /sys/{ProductKey}/{DeviceName}/thing/register
	_ = emqx.Subscribe("/sys/+/+/thing/register", func(topic string, payload []byte) {
		segs := strings.Split(strings.Trim(topic, "/"), "/")
		if len(segs) < 4 {
			return
		}
		productKey := segs[1]
		deviceName := segs[2]
		var body map[string]interface{}
		if err := json.Unmarshal(payload, &body); err != nil {
			return
		}
		
		var d model.Device
		if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
			// 设备不存在，可能是预注册模式，创建新设备
			var p model.Product
			if err := dao.DB.Where("product_key = ?", productKey).First(&p).Error; err != nil {
				return
			}
			// 生成随机 DeviceSecret
			deviceSecret := fmt.Sprintf("%s%08d", productKey[:8], rand.Intn(100000000))
			d = model.Device{
				DeviceName:    deviceName,
				DeviceSN:      fmt.Sprintf("%s_%s", productKey, deviceName),
				DeviceSecret:  deviceSecret,
				ProductKey:    productKey,
				ProductID:     p.ID,
				OwnerID:       1, // 系统用户
				Status:        1,
				NodeType:      "device",
				BindMode:      "product_secret",
			}
			if err := dao.DB.Create(&d).Error; err != nil {
				return
			}
		}
		
		// 返回设备密钥
		response := map[string]interface{}{
			"id":      body["id"],
			"code":    200,
			"message": "success",
			"data": map[string]interface{}{
				"deviceSecret": d.DeviceSecret,
			},
		}
		responseJSON, _ := json.Marshal(response)
		replyTopic := fmt.Sprintf("/sys/%s/%s/thing/register_reply", productKey, deviceName)
		emqx.Publish(replyTopic, response)
		logic.LogMqttMessageAsync(d.DeviceSN, replyTopic, "down", string(responseJSON))
	})

	log.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /sys/+/+/thing/event/property/batch_post, /sys/+/+/thing/event/property/history_post, /ota/device/inform/+/+, /sys/+/+/event/post, /sys/+/+/service/+/reply, /sys/+/+/rrpc/request/+, /sys/+/+/thing/register")
}"""

if old in content:
    content = content.replace(old, new)
    with open('main.go', 'w', encoding='utf-8') as f:
        f.write(content)
    print('Replaced successfully')
else:
    print('Old string not found')
    idx = content.find('log.Println("[EMQX] subscribed Feiyan-style topics:')
    if idx >= 0:
        print('Found at index:', idx)
        print(content[idx:idx+200])