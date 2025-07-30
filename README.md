# OBC Coin API

一个用于创建和发布代币的 API 服务，支持代币模板生成、编译和发布到 Benfen 网络。

## 功能特性

- 🪙 代币模板生成和自定义
- 🔧 自动编译 Move 智能合约
- 🚀 一键发布到 Benfen 网络
- 📝 清晰的参数结构和错误处理

## API 接口

### 1. 添加代币 - `/api/token/add`

创建代币模板并编译生成字节码。

**请求方法：** `POST`

**请求参数：**
```json
{
  "decimal": 8,
  "symbol": "TEST",
  "name": "Test Token",
  "description": "A test token for complete workflow",
  "json": {
    "website": "https://test.com",
    "twitter": "@test"
  }
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "代币添加和编译成功",
  "data": {
    "compile_output": "编译输出信息...",
    "output_file": "/path/to/generated/file.move",
    "request": {...}
  }
}
```

### 2. 发布代币 - `/api/token/publish`

将编译后的代币发布到 Benfen 网络。

**请求方法：** `POST`

**请求参数：**
```json
{
  "sender": "发送者地址",
  "compiled_modules": ["编译后的模块数组"],
  "dependencies": ["依赖项数组"],
  "gas_budget": "Gas 预算"
}
```

## 完整测试流程

### 步骤 1：创建代币

```bash
curl -X POST http://localhost:8080/api/token/add \
  -H "Content-Type: application/json" \
  -d '{
    "decimal": 8,
    "symbol": "TEST",
    "name": "Test Token",
    "description": "A test token for complete workflow",
    "json": {
      "website": "https://test.com",
      "twitter": "@test"
    }
  }'
```

**预期响应：**
- 状态码：200
- 包含编译输出、模块和依赖信息
- 响应时间：约 2-3 秒

### 步骤 2：发布代币

使用步骤 1 返回的编译结果：

```bash
curl -X POST http://localhost:8080/api/token/publish \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "BFC6f0f9a9a72f7d48b8fcbfa09ebb61123d847aaad5760297d68c64795bad514b14a89",
    "compiled_modules": ["从步骤1获取的编译模块"],
    "dependencies": ["从步骤1获取的依赖项"],
    "gas_budget": "5000000000"
  }'
```

**预期响应：**
- 状态码：200
- 来自 Benfen RPC 的发布结果
- 响应时间：约 300-500 毫秒

## 配置说明

服务器配置文件 `config.yaml`：

```yaml
bfc:
  directory: "/path/to/bfc"
  binary_path: "/path/to/bfc/target/debug/bfc"
  token_template_path: "/path/to/token/template"

benfen_rpc:
  url: "https://devrpc2.benfen.org/"
  timeout: 30
  retry_count: 3

server:
  port: 8080
```

## 启动服务

```bash
# 克隆项目
git clone <repository-url>
cd obc_coin_api

# 安装依赖
go mod tidy

# 启动服务
go run .
```

服务启动后将在 `localhost:8080` 监听请求。

## 错误处理

所有接口都包含统一的错误响应格式：

```json
{
  "success": false,
  "message": "错误描述信息"
}
```

常见错误：
- 400：请求参数格式错误
- 500：服务器内部错误（编译失败、网络错误等）

## 技术栈

- **后端框架：** Go + Gorilla Mux
- **智能合约：** Move 语言
- **编译工具：** BFC (Benfen Compiler)
- **网络：** Benfen 区块链网络

## 开发说明

项目结构：
```
├── config.go          # 配置文件处理
├── config.yaml        # 服务配置
├── handlers.go        # API 处理函数
├── main.go           # 服务入口
└── templates/        # 代币模板目录
```

主要功能模块：
- `addToken`: 代币模板生成和编译
- `publishToken`: 代币发布到区块链
- `compileMoveProject`: Move 项目编译
- `processTemplate`: 模板文件处理