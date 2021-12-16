FROM nginx
ADD main.html.template /usr/share/nginx/test/index.html.template
ADD default.conf /etc/nginx/conf.d/default.conf
ADD main.min.js /usr/share/nginx/html/assets/main.min.js
ENV NGINX_ENVSUBST_TEMPLATE_DIR=/usr/share/nginx/test
ENV NGINX_ENVSUBST_OUTPUT_DIR=/usr/share/nginx/html
ENV SERVER_ENDPOINT=http://localhost:8081
ENV INGRESS_PATH=""
EXPOSE 8080