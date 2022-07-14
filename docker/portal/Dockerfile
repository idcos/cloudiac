FROM cloudiac/base-iac-portal:v0.1.7

EXPOSE 9030

COPY configs/config-portal.yml.sample config-portal.yml
COPY configs/dotenv.sample .
COPY build/iac-portal .
COPY build/iac-tool .

CMD  ./iac-portal
