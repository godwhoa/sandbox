sandbox
=======
> An **attempt** at making a language sandbox for programming assignments
>
> Under the hood it uses Docker + seccomp.

## Status
It's currently on-par or even better(non-root user + seccomp profile) than [rust playground](https://github.com/integer32llc/rust-playground).


## TODO

- [x] Run as non-root user
- [x] Limit resources
- [x] No network
- [x] Execution time limit
- [x] Restricted FS
- [x] Drop all capabilities http://man7.org/linux/man-pages/man7/capabilities.7.html
- [x] Set PID limit
- [x] Seccomp profile
- [ ] Audit/Testing
- [ ] Destroy a VM
- [x] Backend
- [ ] Frontend

## Usage
```
# Tested with go 1.10 and Docker 18.03.0-ce
cd backend && go run *.go

cd testfiles && curl -F 'src=@main.c' http://localhost:8080/run
```

## Resources
- https://xtermjs.org/
- https://security.stackexchange.com/questions/107850/docker-as-a-sandbox-for-untrusted-code/107853#107853
- https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities
- https://github.com/integer32llc/rust-playground/blob/master/ui/src/sandbox.rs#L330
- https://www.youtube.com/watch?v=LNMW-38Y8W4
- https://blog.jessfraz.com/post/containers-security-and-echo-chambers/
- https://www.nccgroup.trust/globalassets/our-research/us/whitepapers/2016/april/ncc_group_understanding_hardening_linux_containers-1-1.pdf
- https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux_atomic_host/7/html/container_security_guide/linux_capabilities_and_seccomp
- https://github.com/genuinetools/contained.af/blob/master/seccomp.go