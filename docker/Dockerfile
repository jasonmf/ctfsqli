FROM debian:latest

ENV LISTEN=:8000
ENV TEMPLATE_DIR=/home/ctfsqli/templates
ENV TARGET_FLAG=flag:foobar

RUN useradd -m ctfsqli
RUN chown ctfsqli:ctfsqli /home/ctfsqli
RUN apt update && apt install --no-install-recommends -y default-mysql-server && rm -rf /var/lib/apt/lists/* /var/lib/mysql/ib_logfile*

EXPOSE 8000/tcp

ADD work/ctfsqli /home/ctfsqli/ctfsqli
ADD entrypoint.sh /entrypoint.sh
ADD db.sql /db.sql
ADD templates/ /home/ctfsqli/templates

ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]