FROM mysql:5.7
CMD [ "mysqld", \
  "--character-set-server=utf8mb4", \
  "--collation-server=utf8mb4_unicode_ci", \
  "--sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"]

