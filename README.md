# gocmd

## createmock

```
(defun insert-mock (function-declare)
  (interactive "sFunction Declare: ")
  (let ((command (format "createmock \"%s\"" function-declare)))
    (insert (shell-command-to-string command))))
```

how to use

1. Copy the function declare
2. Move current position
3. M-x insert-mock
4. Yank the function declare

## how to build the binary for raspberry pi

```
$ make all GOOS=linux GOARCH=arm GOARM=6
```

## how to build the binary for AWS Lambda

```
$ make all GOOS=linux
```

## how to deploy AWS Lambda

```
$ aws lambda update-function-code --function-name mindra --zip-file fileb://mindra.zip
```