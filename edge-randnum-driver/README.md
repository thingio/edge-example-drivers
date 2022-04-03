# edge-randnum-driver

这是一个样例项目，用于展示如何基于 edge-device-driver 开发自己的设备驱动程序，该项目实现了随机数生成。

## 使用方法

1. 将 `edge-randnum-driver/etc/resources` 目录下与 `randnum` 相关的产品与设备拷贝至 `edge-device-manager/etc/resources` 的对应目录下；

2. 先启动 edge-device-manager，再启动 edge-randnum-driver；

3. 待驱动初始化完成后，直接通过 MessageBus 进行测试；

4. 测试 HARD-READ 指令：

   ```shell
   # 开启一个终端 订阅设备数据的主题
   mosquitto_sub -h 127.0.0.1 -p 1883 -t DATA/v1/# | grep -v STATUS | grep -v state
   ```

   ```shell
   # 开启另一个终端 模拟 manager 向设备发送读取属性的指令
   # 向协议为 randnum，产品 ID 为 randnum-test01，且 ID 为 randnum-test01 的设备发送 HARD-READ float 属性的指令，（123 表示请求 ID，可随意指定）
   mosquitto_pub -h 127.0.0.1 -p 1883 -t "DATA/v1/DOWN/randnum/randnum-test01/randnum-test01/float/HARD-READ/123" -m ""
   ```

5. 其余指令测试的 TOPIC 格式参考 [物模型操作的 TOPIC 约定](https://github.com/thingio/edge-device-std/blob/main/docs/zh/README.md#%E7%89%A9%E6%A8%A1%E5%9E%8B%E6%93%8D%E4%BD%9C)。