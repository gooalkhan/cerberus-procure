@echo off
REM Windows Batch wrapper for make.ps1
powershell -ExecutionPolicy Bypass -File "%~dp0make.ps1" %*
