@echo off
mkdir "%USERPROFILE%\.config\gottem"
copy gottem_windows_amd64.exe "%USERPROFILE%\AppData\Local\Microsoft\WindowsApps\gottem.exe"
echo Gottem v1.2.0 installed successfully!
