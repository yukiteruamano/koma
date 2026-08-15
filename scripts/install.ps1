# WAS NOT TESTED

$release_url = "https://api.github.com/repos/yukiteruamano/koma/releases"
$tag = (Invoke-WebRequest -Uri $release_url -UseBasicParsing | ConvertFrom-Json)[0].tag_name
$version = $tag.substring(1)
$loc = "$HOME\AppData\Local\koma"
$url = ""
$arch = $env:PROCESSOR_ARCHITECTURE
$releases_api_url = "https://github.com/yukiteruamano/koma/releases/download/$tag/koma_${version}_Windows"

if ($arch -eq "AMD64")
{
    $url = "${releases_api_url}_x86_64.zip"
}
elseif ($arch -eq "x86")
{
    $url = "${releases_api_url}_i386.zip"
}
elseif ($arch -eq "arm64")
{
    $url = "${releases_api_url}_arm64.zip"
}

if (Test-Path -path $loc)
{
    Remove-Item $loc -Recurse -Force
}

Write-Host "Installing koma version $tag" -ForegroundColor DarkCyan

Invoke-WebRequest $url -outfile koma.zip

Expand-Archive koma.zip

New-Item -ItemType "directory" -Path $loc

Move-Item -Path koma\koma.exe -Destination $loc

Remove-Item koma* -Recurse -Force

[System.Environment]::SetEnvironmentVariable("Path", $Env:Path + ";$loc", [System.EnvironmentVariableTarget]::User)

if (Test-Path -path $loc)
{
    Write-Host "Koma version $tag installed successfully" -ForegroundColor Green
}
else
{
    Write-Host "Download failed" -ForegroundColor Red
    Write-Host "Please try again later" -ForegroundColor Red
}
