# SMTPServerToWebHook
An idea I got from a camera bridge thing


## Config
Modify the config.yaml for personalization.

## Notice
This is not secure. Use at own risk

## Testing
```
apt install sendemail


sendEmail -xu user -xp password -u "test_email" -m "test" -s "127.0.0.1:2525" -f me@email.com -t target@email.com

```