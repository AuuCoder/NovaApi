# nginx vhost

NovaAPI 在 `92.119.167.65` 上的 nginx 反代配置。Cloudflare 在前面做 Full SSL,nginx 监听 443 直接服务。

## 文件

- `muyuai-novapi.conf` — 主 vhost,监听 `muyuai.top`,反代到 `127.0.0.1:8090`(sub2api 容器)
- `muyuai-www-redirect.conf` — `www.muyuai.top` 永久 301 到 apex,统一 cookie 域

## 部署

```bash
# 假设证书已放在 /etc/nginx/certs/muyuai.top.{crt,key}
sudo cp deploy/nginx/muyuai-novapi.conf       /etc/nginx/sites-available/muyuai-novapi
sudo cp deploy/nginx/muyuai-www-redirect.conf /etc/nginx/sites-available/muyuai-www-redirect

sudo ln -sfn /etc/nginx/sites-available/muyuai-novapi       /etc/nginx/sites-enabled/muyuai-novapi
sudo ln -sfn /etc/nginx/sites-available/muyuai-www-redirect /etc/nginx/sites-enabled/muyuai-www-redirect

sudo nginx -t && sudo systemctl reload nginx
```

## 为什么需要 www→apex 跳转

OAuth state cookie 是 host-only。如果用户从 `www.muyuai.top` 进入触发 `/start`(cookie 写在 www),
而 LinuxDo / Google 等 provider 的 `redirect_uri` 配的是 apex `muyuai.top`,
回调时浏览器视为不同域,cookie 完全不发,后端 state 校验必失败(`invalid_state`)。

把 www 统一 301 到 apex 后,所有用户的 OAuth 流程都在同一个域上完成,问题消失。
