@echo off
REM Magic-ORM 代码生成器构建脚本 (Windows)

echo.
echo 🔨 开始构建 Magic-ORM 代码生成器...
echo.

REM 设置版本号
set VERSION=1.0.0
set BINARY_NAME=gendb.exe

REM 创建 bin 目录
if not exist bin mkdir bin

REM 构建当前平台的版本
echo 📦 构建 Windows 版本...
go build -o bin\%BINARY_NAME% -ldflags="-s -w" ./cmd/gendb

echo.
echo ✅ 构建完成!
echo.
echo 📍 可执行文件位置:
echo    - bin\%BINARY_NAME%
echo.
echo 💡 使用方法:
echo    1. 添加到 PATH: set PATH=%%PATH%%;%%CD%%\bin
echo    2. 直接使用：gendb user product
echo    3. 查看帮助：gendb -h
echo.
echo 🪟 或者将 bin 目录添加到系统环境变量 PATH 中
echo.
