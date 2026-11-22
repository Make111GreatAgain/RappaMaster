## RappaMaster

### 一、环境准备

- FISCO-BCOS

    - 搭建fiscobcos3.0版本：参考 [https://fisco-bcos-documentation.readthedocs.io/zh-cn/latest/docs/installation.html#id2](https://fisco-bcos-doc.readthedocs.io/zh-cn/latest/docs/quick_start/air_installation.html)
    - 配置文件：将fisco目录下 127.0.0.1/sdk文件夹下的ca.crt、sdk.crt、sdk.key替换到文件夹相应位置
    - 启动FISCO-BCOS：运行 fisco/nodes/127.0.0.1/start_all.sh

- 异步上链

    - FISCO-BCOS GO-SDK

        - go-sdk需要依赖csdk的动态库

          ```sh
          # 下面的脚本帮助用户下载bcos-c-sdk的动态库到/usr/local/lib目录下
          ./tools/download_csdk_lib.sh
          
          # 下载完成时需要设置动态库的搜索路径
          export LD_LIBRARY_PATH=${PWD}/lib # 上面设置到了/usr/local/lib下因此是export LD_LIBRARY_PATH=/usr/local/lib
          ```

### 二、HTTP测试客户端 (http_client_test.py)

- 功能概述：这个脚本提供了一个简单的 HTTP 客户端，能够发送请求与`Master`中的`HttpEngine`交互。它通过 `requests` 库模拟与服务端的 HTTP 通信，通过shell的形式提供交互，目前提供以下命令

    - `create`:创建一个合成任务

        - 请求方式: POST

        - URL: `http://127.0.0.1:8080/create`

        - 请求体示例

          ```json
          {
              "model": "CTGAN",
              "params": {
                  "condition_column": "native-country",
                  "condition_value": "United-States"
              },
              "size": 50,
              "isReliable": true
          }
          
          ```

        - 会返回一个`taskID`，可修改代码中对`task`的请求

    - `task`：查询一个合成任务

        - 请求方式：GET

        - URL: `http://127.0.0.1:8080/oracle`

        - 请求体示例

          ```json
          {
              "query": "EvidencePreserveEpochIDQuery",
              "epochID": 8
          }
          ```

        - 会返回一个合成任务在`oracle`中的记录细节

    - `epoch`：查询一个`epoch`

        - 请求方式: `GET`

        - URL: `http://127.0.0.1:8080/oracle`

        - 请求体示例

          ```json
          {
              "query": "EvidencePreserveEpochIDQuery",
              "epochID": 8
          }
          ```

        - 会返回一个epoch在`oracle`的记录细节

### 三、运行方式

- 先运行`Executor`

  ```shell
  # 在对应路径下
  python3 main.py
  # 等待所有模型加载完全
  ```

- 运行`Master`

  ```shell
  # 在对应路径下
  bash run.sh
  ```

- 运行HTTP客户端

  ```shell
  python3 http_client_test.py
  
  # 在shell中输入
  create # 新建一个合成任务，这样Master才会开始调度
  
  
  # 可选，运行过程中或运行完成后
  epoch # 查询epoch
  task # 查询task
  ```
## Rappa部署文档

### 一、安装依赖

- 操作系统：Ubuntu: 18.04以上

- 依赖库如gcc安装

```shell
# 若出现permission denied，请加上sudo
apt update

apt install -y g++ cmake git build-essential autoconf texinfo flex patch bison libgmp-dev zlib1g-dev automake libtool python3-pip libcairo2-dev pkg-config curl wget
```

- `openssl-3.x`安装

    - 方法一：参考教程自行安装，参考链接：https://blog.csdn.net/m0_65803902/article/details/142686409

    - 方法二：根据提供的压缩包安装`openssl-3.4.0`

      ```shell
      # 若本地已安装了不合条件的openssl，请参考方法一中部分进行去除或升级,以下命令仅限于安装openssl-3.4.0
      tar xvf openssl-3.4.0.tar.gz
      cd openssl-3.4.0
      ./config --prefix=/usr/local/openssl
      make && make install# 若有多核，可以使用make -j{并行核数}，如make -j8将使用8个逻辑核
      
      cd /usr/local/openssl/bin
      ldd openssl # 验证软链接，如果有not found，参考上面的链接
      
      # 确认ldd openssl后
      vim /etc/profile
      # 在文件末尾加入以下两行
      export OPENSSL=/usr/local/openssl/bin
      export PATH=$OPENSSL:$PATH:$HOME/bin
      # vim完成后
      source /etc/profile
      openssl version #输出版本号3.4.0说明安装成功
      ```



- 安装`python`

  目前项目运行的是`python3.8`,若为`3.10`以上或`3.8`以下可能因为python的兼容问题出现报错，推荐使用`python3.8`

    - 方法一：参考教程自行安装`python3.8`，参考链接：https://blog.csdn.net/qq_62204036/article/details/142101925



- 安装`Golang`

    - 方法一：参考教程自行安装`Golang`，版本需要1.20以上，参考链接：https://blog.csdn.net/lza20001103/article/details/145149141

    - 方法二：根据提供的压缩包安装Golang-1.22.8

      ```shell
      # 进入到压缩包的目录
      tar xvf go1.22.8.linux-amd64.tar.gz {解压路径}
      
      vim /etc/profile
      # 在文件末尾加入以下两行
      export GOBIN={这里替换为解压后go文件夹的路径}/go/bin
      export PATH=$PATH:$GOBIN
      
      # vim完成后
      source /etc/profile
      go version # 输出go版本号说明安装成功
      
      export GOPROXY=https://goproxy.cn,direct # 设置GO PROXY
      ```

- 安装FISCO-BCOS GO-SDK

    - ` Go-sdk`需要依赖csdk的动态库

  ```shell
  # 下面的脚本帮助用户下载bcos-c-sdk的动态库到/usr/local/lib目录下
  ./tools/download_csdk_lib.sh
  
  # 下载完成时需要设置动态库的搜索路径
  export LD_LIBRARY_PATH=${PWD}/lib # 上面设置到了/usr/local/lib下因此是export LD_LIBRARY_PATH=/usr/local/lib
  
  # 注： 如果没有/usr/local/lib这个文件夹，那么打开tools/download_csdk_lib.sh文件，修改其中的install_path为需要的路径，然后对应修改下面export的路径
  
  ```

### 二、FISCO BCOS 3.x 部署流

- 创建操作目录 `fisco`，下载安装脚本

  ```sh
  # 在某个目录下创建操作目录fisco
  mkdir -p fisco && cd fisco
  
  # 下载建链脚本（可能因为网络问题下载失败，请尝试下面一个命令）
  curl -#LO https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.11.0/build_chain.sh && chmod u+x build_chain.sh
  
  # Note: 若访问git网速太慢，可尝试如下命令下载建链脚本:
  curl -#LO https://gitee.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.11.0/build_chain.sh && chmod u+x build_chain.sh
  ```

- 搭建4节点联盟链

    - 在fisco目录下执行下面的指令，生成一条单群组4节点的FISCO链:

      ```sh
      cd fisco
      bash build_chain.sh -l 127.0.0.1:4 -p 30300,20200
      ```

      运行成功后会输出 `All completed` :

        ```
      [INFO] Generate ca cert successfully!
      Processing IP:127.0.0.1 Total:4
      writing RSA key
      [INFO] Generate ./nodes/127.0.0.1/sdk cert successful!
      writing RSA key
      [INFO] Generate ./nodes/127.0.0.1/node0/conf cert successful!
      writing RSA key
      [INFO] Generate ./nodes/127.0.0.1/node1/conf cert successful!
      writing RSA key
      [INFO] Generate ./nodes/127.0.0.1/node2/conf cert successful!
      writing RSA key
      [INFO] Generate ./nodes/127.0.0.1/node3/conf cert successful!
      [INFO] Downloading get_account.sh from https://gitee.com/FISCO-BCOS/console/raw/master/tools/get_account.sh...
      ############################################################################################################################################################### 100.0%
      [INFO] Admin account: 0x4c7239cfef6d41b7322c1567f082bfc65c69acc5
      [INFO] Generate uuid success: 167A2233-5444-4CA4-8792-C8E68130D5FC
      [INFO] Generate uuid success: CC117CAF-224C-4940-B548-6DED31D24B18
      [INFO] Generate uuid success: 16B5E4BD-51C1-416E-BF44-6D1BB05F7666
      [INFO] Generate uuid success: 60DD77F2-F3A5-49F2-8C7F-8151E8823C6D
      ==============================================================
      [INFO] GroupID              : group0
      [INFO] ChainID              : chain0
      [INFO] fisco-bcos path      : bin/fisco-bcos
      [INFO] Auth mode            : false
      [INFO] Start port           : 30300 20200
      [INFO] Server IP            : 127.0.0.1:4
      [INFO] SM model             : false
      [INFO] enable HSM           : false
      [INFO] Output dir           : ./nodes
      [INFO] All completed. Files in ./nodes
        ```

- 启动FISCO BCOS链

    - 启动所有节点（Note：运行RappaMaster前FISCO BCOS链应处于启动状态）

      ```sh
      bash /root/fisco/nodes/127.0.0.1/start_all.sh
      ```

      成功后会输出如下信息。否则请使用 `netstat -an |grep tcp` 检查机器 `30300~30303, 20200~20203` 端口是否被占用。

        ```
      try to start node0
      try to start node1
      try to start node2
      try to start node3
      node3 start successfully pid=36430
      node2 start successfully pid=36427
      node1 start successfully pid=36433
      node0 start successfully pid=36428
        ```

    - 检查节点进程是否成功启动

      ```sh
      ps aux |grep -v grep |grep fisco-bcos
      ```

      正常情况下会有类似下面的输出； 如果进程数不为4，则进程没有启动（一般是端口被占用导致的）

        ```
      fisco        35249   7.1  0.2  5170924  57584 s003  S     2:25下午   0:31.63 /home/fisco/nodes/127.0.0.1/node1/../fisco-bcos -c config.ini -g config.genesis
      fisco        35218   6.8  0.2  5301996  57708 s003  S     2:25下午   0:31.78 /home/fisco/nodes/127.0.0.1/node0/../fisco-bcos -c config.ini -g config.genesis
      fisco        35277   6.7  0.2  5301996  57660 s003  S     2:25下午   0:31.85 /home/fisco/nodes/127.0.0.1/node2/../fisco-bcos -c config.ini -g config.genesis
      fisco        35307   6.6  0.2  5301996  57568 s003  S     2:25下午   0:31.93 /home/fisco/nodes//127.0.0.1/node3/../fisco-bcos -c config.ini -g config.genesis
        ```

- 更新 RappaMaster 区块链配置文件

    - 将 `fisco/nodes/127.0.0.1/sdk` 目录下的 `ca.crt、sdk.crt、sdk.key` 替换到 `RappaMaster/ChainUpper` 位置

      ```sh
      # 请将 /root/rappa/RappaMaster/ChainUpper 替换为该文件夹的真实路径
      cd /root/fisco/nodes/127.0.0.1/sdk
      cp ca.crt sdk.crt sdk.key /root/rappa/RappaMaster/ChainUpper/
      ```

其他：

- 关闭FISCO BCOS命令

```sh
bash /root/fisco/nodes/127.0.0.1/stop_all.sh
```

### 三、RappaMaster

- 代码下载

    - 方法一：根据压缩包解压得到

    - 方法二：通过github下载（推荐），方便代码更新

      ```shell
      git clone https://github.yuuza.net/ZhmYe/RappaMaster.git
      ```

- 创建日志目录文件

  ```shell
  mkdir logs
  # 代码运行的所有中间输出都会以日志形式记录在logs/下面形成一个日志文件.log
  # 若有运行错误，可参考日志文件
  ```



- 运行代码

  ```shell
  bash run.sh
  ```



### 四、RappaExecutor

- 代码下载

    - 方法一：根据压缩包解压得到

    - 方法二：通过github下载（推荐），方便代码更新

      ```shell
      git clone https://github.yuuza.net/ZhmYe/RappaExecutor.git
      ```

- 依赖库安装

  ```shell
  pip install -r requirement.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
  
  # 安装torch-scatter，需根据要使用的device环境cpu/gpu进行安装,以下是cpu示例，gpu示例见torch-scatter官网
  pip install torch-scatter -f https://data.pyg.org/whl/torch-2.4.0+cpu.html -i https://pypi.tuna.tsinghua.edu.cn/simple
  
  # dgl安装
  pip install dgl -f https://data.dgl.ai/wheels/torch-2.4/repo.html -i https://pypi.tuna.tsinghua.edu.cn/simple
  
  
  ```

- 创建日志目录文件

  ```shell
  mkdir logs
  # 代码运行的所有中间输出都会以日志形式记录在logs/下面形成一个日志文件.log
  # 若有运行错误，可参考日志文件
  ```

- 运行代码

  ```shell
  python main,py # 启动一个节点，批量启动节点见第五节
  ```



### 五、启动Rappa

- 



### 空 epoch 明细与历史清理

发布新代码前确认 MySQL 有复合索引 `idx_epoch_ref (rf, epoch_id)`，用于最大编号与最近列表查询。如果开启了 GORM 自动迁移，可由新版本创建。否则在部署窗口核对索引后执行 `ALTER TABLE dev_references ADD INDEX idx_epoch_ref (rf, epoch_id);`。发布新代码后先以默认配置运行兼容查询。确认该部署使用恢复模式（`IsRecovery: true`，旧有非恢复模式启动会清表），再在部署配置中设置 `"SkipEmptyEpochDetails": true` 并重启。心跳、上链交易和 MySQL 交易引用仍会保留。空 epoch 的明细由查询临时重建。直接查询 `dev_epoches` 的报表会看到稀疏数据，需要先核查这些调用方。需要停用稀疏写入时可关闭该配置并继续运行兼容版本。

清理默认只预览，报告超过 30 天的候选行和 InnoDB 估算的表大小。运行前确认配置指向目标库。

```sh
go run ./tools/epoch-cleanup -config config.json
go run ./tools/epoch-cleanup -config config.json -apply -max-batches 10
```

每批最多删除 1,000 条主表明细。只有六列都是预期的空 JSON 且交易引用匹配才会删除。保留交易回执、异常记录和链上历史。首次执行前在备份或测试库抽样，用原 ID 和交易哈希对比详情，核验最新列表和交易统计。首次清理确认后，可以将构建出的工具（`go build -o epoch-cleanup ./tools/epoch-cleanup`）用绝对路径加入每日任务，显式传入 `-config` 和 `-apply`。InnoDB 删除后通常复用内部空间，磁盘文件不一定立即缩小。
