@echo off
chcp 65001 >nul
cls
echo.
echo ========================================
echo   Magic-ORM 代码生成器
echo ========================================
echo.

go run ./cmd/gendb %*

echo.
echo 按任意键退出...
pause >nul
