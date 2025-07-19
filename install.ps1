$ErrorActionPreference = "Stop"  # Stop on errors


$progName = "Mujamalat"
$programDataPath = [System.Environment]::GetFolderPath('ProgramFiles')
$progPath = Join-Path -Path $programDataPath -ChildPath $progName

# Define the URL of the file and the path where you want to save it
$icoUrl = "https://raw.githubusercontent.com/wizsk/mujamalat/refs/heads/main/pub/fav.ico"
$url = "https://github.com/wizsk/mujamalat/releases/latest/download"
$sysTmp = [System.IO.Path]::GetTempPath()
$file = "mujamalat_windows_x86_64.zip"
$downFilePath = Join-Path -Path $sysTmp -ChildPath $file
$icoPath = Join-Path -Path $progPath -ChildPath "fav.ico"

# Check if the folder exists
if (-Not (Test-Path -Path $progPath)) {
    New-Item -Path $progPath -ItemType Directory
    Write-Host "Directory created: $progPath"
} else {
    Write-Host "Directory already exists: $progPath"
}

try {
    Write-Host "Downloading: $url/$file"
    Invoke-WebRequest -Uri "$url/$file" -OutFile $downFilePath
    Write-Host "Downloading: $icoUrl"
    Invoke-WebRequest -Uri "$icoUrl" -OutFile $icoPath
}
catch {
    Write-Host "Err: Downlaoing failed! bye"
    return
}


$exFolder = Join-Path -Path $sysTmp -ChildPath $progName

try {
    Write-Host "Extracting: $downFilePath"
    Expand-Archive -Path $downFilePath -DestinationPath $exFolder -Force
}
catch {
    Write-Host "err: Extracting failed $downFilePath bye"
}


try {
    Write-Host "Copying exe to $progPath"
    $hwExePath = Join-Path -Path $exFolder -ChildPath "mujamalat.exe"
    Copy-Item -Path $hwExePath -Destination $progPath
}
catch {
    Write-Host "Err: Copying $hwExePath to $progPath failed. bye!"
    return
}

Write-Host "Adding $progPath to SYSTEM PATH"
$systemPath = [System.Environment]::GetEnvironmentVariable("PATH", [System.EnvironmentVariableTarget]::Machine)
# Check if the bin folder is already in the PATH
if ($systemPath -notlike "*$progPath*") {
    $newSystemPath = "$systemPath;$progPath"
    [System.Environment]::SetEnvironmentVariable("PATH", $newSystemPath, [System.EnvironmentVariableTarget]::Machine)
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", [System.EnvironmentVariableTarget]::Machine)
}


# Removing files
try {
    Remove-Item -Path "$exFolder" -Recurse
    Remove-Item -Path $downFilePath

}
catch {}


# adding desktop shorcut
$arguments = ""
$desktopPath = [System.IO.Path]::Combine($env:USERPROFILE, "Desktop")
$shortcutPath = [System.IO.Path]::Combine($desktopPath, "$progName.lnk")
$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
$hwExeFullPath = Join-Path -Path $progPath -ChildPath "hw.exe"
$shortcut.TargetPath = $hwExeFullPath
$shortcut.Arguments = $arguments
$shortcut.IconLocation = $icoPath
$shortcut.Save()

# adding desktop shartmenu for search
$arguments = ""
$startMenuPath = [System.IO.Path]::Combine($env:ProgramData, "Microsoft\Windows\Start Menu\Programs")
$shortcutPath = [System.IO.Path]::Combine($startMenuPath, "$progName.lnk")
$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = $hwExeFullPath
$shortcut.Arguments = $arguments
$shortcut.IconLocation = $icoPath
$shortcut.Save()
Write-Host "Shortcut created in the Start Menu for all users. It will appear in Windows Search."

$shortcutPath = [System.IO.Path]::Combine($progPath, "$progName.lnk")
$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
# Set the target executable and arguments
$hwExeFullPath = Join-Path -Path $progPath -ChildPath "hw.exe"
$shortcut.TargetPath = $hwExeFullPath
$shortcut.Arguments = $arguments
$shortcut.IconLocation = $icoPath
# Save the shortcut
$shortcut.Save()

Write-Host ""
Write-Host "Installation compleaded! Now run 'mujamalat'"

