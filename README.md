<h1 align="center">
    <img src="https://img.alicdn.com/imgextra/i2/O1CN01NwxLDd20nxfGBjxmZ_!!6000000006895-2-tps-960-290.png" alt="AIGateway" width="240" height="72.5">
  <br>
  Gateway Console for AIGateway
</h1>

AIGateway Console 用于管理 AIGateway 的配置规则及其他开箱即用的能力集成，首个可用版本考虑基于 kubernetes 部署环境，预期包含服务管理、路由管理、域名管理等基础能力。
后续规划逐步迭代可观测能力、插件能力、登录管理能力。

## 前置介绍

此项目包含前端（NodeJS）、后端（GoFrame）两个部分，前端（frontend）部分在构建完成后会随着后端代码（goframe）一起部署。

## 本地启动

### 前端项目

#### 第一步、配置 Node 环境

注：建议 Node 版本选择长期稳定支持版本 16.18.1 及以上

#### 第二步、安装依赖

```bash
cd frontend && npm install
```

#### 第三步、本地启动

```bash
npm start
```

#### 第四步、打包

```bash
npm run build
#打包生成文件 frontend/build
```

### 后端项目

#### 第一步、配置 Go 环境

注：建议 Go 版本选择 1.23 及以上。

#### 第二步、说明当前目录结构

```bash
backend              # 新的 GoFrame 后端
backend-java-legacy  # 原 Java/Spring Boot 后端迁移参考基线
```

#### 第三步、编译 & 测试

```bash
cd backend && ./build.sh
```

#### 第四步、部署 & 启动

```bash
cd backend && ./start.sh
```

#### 第五步、访问

主页，默认 8080 端口

```html
http://localhost:8080/landing
```

可以通过以下方法开启 Swagger UI，并通过访问 Swagger 页面了解 API 情况。

**Swagger UI URL：**
```html
http://localhost:8080/swagger
```

## 开发规范

### 后端项目
