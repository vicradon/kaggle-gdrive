# gdrive

A minimal CLI tool to upload files and directories to Google Drive using a service account.

## Usage

```bash
export GDRIVE_SA_KEY='<contents of your service-account JSON key>'

# Upload a directory
gdrive -from /kaggle/working/runs/train -to "/drive/My Drive/PetCrib"

# Upload a single file
gdrive -from /kaggle/working/runs/train/best.pt -to "/drive/My Drive/PetCrib"
```

The `-to` path is flexible — all of these resolve to the same destination:

```
"/drive/My Drive/PetCrib"
"My Drive/PetCrib"
"PetCrib"
```

Nested folder paths are created automatically if they don’t exist:

```bash
gdrive -from ./results -to "/drive/My Drive/Experiments/Run42/weights"
```

## Setup

1. Create a Google Cloud service account and download its JSON key.
1. Share your target Drive folder with the service account’s `client_email`.
1. Store the JSON key contents in a `GDRIVE_SA_KEY` environment variable
   (or a Kaggle secret with the same name).

## Releases

Pre-built binaries for Linux, macOS, and Windows are published automatically
on every push to `main`. Download from the [Releases](../../releases) page.

|Binary                    |Platform                     |
|--------------------------|-----------------------------|
|`gdrive-linux-amd64`      |Linux x86-64 (Kaggle default)|
|`gdrive-linux-arm64`      |Linux ARM64                  |
|`gdrive-darwin-amd64`     |macOS Intel                  |
|`gdrive-darwin-arm64`     |macOS Apple Silicon          |
|`gdrive-windows-amd64.exe`|Windows x86-64               |
