FROM cloudiac/iac-web:v0.3.0-nginx-error
ADD iac_nginx.conf /etc/nginx/conf.d/iac.conf

