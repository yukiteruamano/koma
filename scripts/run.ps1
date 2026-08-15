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

Write-Host "Downloading Koma version $tag" -ForegroundColor DarkCyan

# download koma to temp folder
$zip = "$env:TEMP\koma.zip"
Invoke-WebRequest -Uri $url -OutFile $zip

# extract koma at temp folder
Expand-Archive -Path $zip -DestinationPath $env:TEMP

# run koma binary from the unzipped folder
$bin = "$env:TEMP\koma\koma.exe"
Start-Process $bin
