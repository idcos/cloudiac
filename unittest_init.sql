-- MySQL dump 10.13  Distrib 8.0.25, for macos11.3 (x86_64)
--
-- Host: 127.0.0.1    Database: iac4
-- ------------------------------------------------------
-- Server version	8.0.25

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `iac_bill`
--

DROP TABLE IF EXISTS `iac_bill`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_bill` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `vg_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `product_code` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `instance_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `instance_config` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `pretax_amount` float NOT NULL,
  `region` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `currency` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cycle` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `provider` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_bill`
--

LOCK TABLES `iac_bill` WRITE;
/*!40000 ALTER TABLE `iac_bill` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_bill` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_bill_data`
--

DROP TABLE IF EXISTS `iac_bill_data`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_bill_data` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `provider` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `instance_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `attrs` json DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_bill_data`
--

LOCK TABLES `iac_bill_data` WRITE;
/*!40000 ALTER TABLE `iac_bill_data` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_bill_data` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_ct_resource_map`
--

DROP TABLE IF EXISTS `iac_ct_resource_map`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_ct_resource_map` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `resource_account_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源账号ID',
  `ct_service_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Runner Service ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__resource_account_id__ct_service_id` (`resource_account_id`,`ct_service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_ct_resource_map`
--

LOCK TABLES `iac_ct_resource_map` WRITE;
/*!40000 ALTER TABLE `iac_ct_resource_map` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_ct_resource_map` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_env`
--

DROP TABLE IF EXISTS `iac_env`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_env` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `status` enum('active','failed','inactive') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `task_status` enum('','approving','running') COLLATE utf8mb4_unicode_ci DEFAULT '',
  `archived` tinyint(1) DEFAULT '0',
  `step_timeout` bigint DEFAULT '3600' COMMENT '部署超时',
  `one_time` tinyint(1) DEFAULT '0',
  `deploying` tinyint(1) NOT NULL DEFAULT '0',
  `tags` text COLLATE utf8mb4_unicode_ci,
  `state_path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tf_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `play_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `playbook` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `runner_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `runner_tags` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `revision` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `key_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `workdir` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `last_task_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_res_task_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_scan_task_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `auto_approval` tinyint(1) DEFAULT '0',
  `stop_on_violation` tinyint(1) DEFAULT '0',
  `ttl` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '0',
  `auto_destroy_at` datetime DEFAULT NULL,
  `auto_destroy_task_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `triggers` text COLLATE utf8mb4_unicode_ci,
  `retry_number` int DEFAULT '3',
  `retry_delay` int DEFAULT '5',
  `retry_able` tinyint(1) DEFAULT '0',
  `extra_data` json DEFAULT NULL,
  `callback` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `cron_drift_express` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `auto_repair_drift` tinyint(1) DEFAULT '0',
  `open_cron_drift` tinyint(1) DEFAULT '0',
  `next_drift_task_time` datetime DEFAULT NULL,
  `policy_enable` tinyint(1) DEFAULT '0',
  `locked` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__project__env__name` (`project_id`,`name`,`deleted_at_t`),
  KEY `idx_iac_env_deleted_at_t` (`deleted_at_t`),
  KEY `idx_iac_env_last_res_task_id` (`last_res_task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_env`
--

LOCK TABLES `iac_env` WRITE;
/*!40000 ALTER TABLE `iac_env` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_env` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_key`
--

DROP TABLE IF EXISTS `iac_key`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_key` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `org_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织ID',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密钥名称',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密钥内容',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org__name` (`org_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_key`
--

LOCK TABLES `iac_key` WRITE;
/*!40000 ALTER TABLE `iac_key` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_key` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_notification`
--

DROP TABLE IF EXISTS `iac_notification`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_notification` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织ID',
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目ID',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `type` enum('email','webhook','wechat','slack','dingtalk') COLLATE utf8mb4_unicode_ci DEFAULT 'email' COMMENT '通知类型',
  `secret` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'dingtalk加签秘钥',
  `url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回调url',
  `user_ids` text COLLATE utf8mb4_unicode_ci COMMENT '用户ID',
  `creator` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_notification`
--

LOCK TABLES `iac_notification` WRITE;
/*!40000 ALTER TABLE `iac_notification` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_notification` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_notification_event`
--

DROP TABLE IF EXISTS `iac_notification_event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_notification_event` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `event_type` enum('task.failed','task.complete','task.approving','task.running','task.crondrift') COLLATE utf8mb4_unicode_ci DEFAULT 'task.running' COMMENT '事件类型',
  `notification_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_notification_event`
--

LOCK TABLES `iac_notification_event` WRITE;
/*!40000 ALTER TABLE `iac_notification_event` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_notification_event` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_operation_log`
--

DROP TABLE IF EXISTS `iac_operation_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_operation_log` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `username` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_addr` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `operation_at` datetime DEFAULT NULL,
  `operation_url` text COLLATE utf8mb4_unicode_ci,
  `operation_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `operation_info` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `operation_status` bigint DEFAULT NULL,
  `desc` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_operation_log`
--

LOCK TABLES `iac_operation_log` WRITE;
/*!40000 ALTER TABLE `iac_operation_log` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_operation_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_org`
--

DROP TABLE IF EXISTS `iac_org`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_org` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织名称',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '组织描述',
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT '组织状态',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人',
  `runner_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_demo` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_org`
--

LOCK TABLES `iac_org` WRITE;
/*!40000 ALTER TABLE `iac_org` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_org` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_policy`
--

DROP TABLE IF EXISTS `iac_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_policy` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '组织ID',
  `group_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '策略组ID',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` text COLLATE utf8mb4_unicode_ci COMMENT '名称',
  `rule_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'rego规则名称',
  `reference_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略ID',
  `revision` bigint DEFAULT '1' COMMENT '版本',
  `enabled` tinyint(1) DEFAULT '1' COMMENT '是否全局启用',
  `fix_suggestion` text COLLATE utf8mb4_unicode_ci COMMENT '策略修复建议',
  `severity` enum('high','medium','low') COLLATE utf8mb4_unicode_ci DEFAULT 'medium' COMMENT '严重性',
  `policy_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '云商类型',
  `resource_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源类型',
  `tags` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签',
  `rego` text COLLATE utf8mb4_unicode_ci COMMENT 'rego脚本',
  PRIMARY KEY (`id`),
  KEY `idx_iac_policy_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_policy`
--

LOCK TABLES `iac_policy` WRITE;
/*!40000 ALTER TABLE `iac_policy` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_policy` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_policy_group`
--

DROP TABLE IF EXISTS `iac_policy_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_policy_group` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '组织ID',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人ID',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略组名称',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `enabled` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `source` enum('vcs','registry') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '来源：VCS/Registry',
  `vcs_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'VCS ID',
  `repo_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'VCS 仓库ID',
  `git_tags` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Git 版本标签："v1.0.0"',
  `branch` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分支',
  `commit_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `use_latest` tinyint(1) DEFAULT '0' COMMENT '是否跟踪最新版本，如果从分支导入，默认为true',
  `version` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `dir` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '/' COMMENT '策略组目录，默认为根目录：/',
  `label` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '策略组标签，多个值以 , 分隔',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__name` (`name`,`deleted_at_t`),
  KEY `idx_iac_policy_group_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_policy_group`
--

LOCK TABLES `iac_policy_group` WRITE;
/*!40000 ALTER TABLE `iac_policy_group` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_policy_group` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_policy_rel`
--

DROP TABLE IF EXISTS `iac_policy_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_policy_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织',
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '项目ID',
  `group_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '策略组ID',
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '云模板ID',
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境ID',
  `scope` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '绑定范围',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__group__tpl__env` (`group_id`,`tpl_id`,`env_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_policy_rel`
--

LOCK TABLES `iac_policy_rel` WRITE;
/*!40000 ALTER TABLE `iac_policy_rel` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_policy_rel` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_policy_result`
--

DROP TABLE IF EXISTS `iac_policy_result`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_policy_result` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织',
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '项目ID',
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '云模板ID',
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '环境ID',
  `task_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务ID',
  `policy_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略ID',
  `policy_group_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略组ID',
  `start_at` datetime DEFAULT NULL COMMENT '开始时间',
  `status` enum('passed','violated','suppressed','pending','failed') COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '状态',
  `message` text COLLATE utf8mb4_unicode_ci COMMENT '失败原因',
  `rule_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '策略名称',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '策略描述',
  `rule_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则ID(策略ID)',
  `severity` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '严重程度',
  `category` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类（策略组名称）',
  `comment` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '跳过说明',
  `resource_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源名称',
  `resource_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源类型',
  `module_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '模块名称',
  `file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源码文件名',
  `plan_root` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源码文件夹',
  `line` bigint DEFAULT NULL COMMENT '错误资源源码行号',
  `source` text COLLATE utf8mb4_unicode_ci COMMENT '错误源码',
  PRIMARY KEY (`id`),
  KEY `idx_iac_policy_result_task_id` (`task_id`),
  KEY `idx_iac_policy_result_start_at` (`start_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_policy_result`
--

LOCK TABLES `iac_policy_result` WRITE;
/*!40000 ALTER TABLE `iac_policy_result` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_policy_result` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_policy_suppress`
--

DROP TABLE IF EXISTS `iac_policy_suppress`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_policy_suppress` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织ID',
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '项目ID',
  `target_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标ID',
  `target_type` enum('env','template','policy') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '屏蔽目标类型',
  `policy_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略ID',
  `reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '屏蔽说明',
  `type` enum('policy','source') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '屏蔽类型',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__policy__target` (`target_id`,`policy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_policy_suppress`
--

LOCK TABLES `iac_policy_suppress` WRITE;
/*!40000 ALTER TABLE `iac_policy_suppress` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_policy_suppress` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_project`
--

DROP TABLE IF EXISTS `iac_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_project` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `creator_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT '状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org__project__name` (`org_id`,`name`,`deleted_at_t`),
  KEY `idx_iac_project_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_project`
--

LOCK TABLES `iac_project` WRITE;
/*!40000 ALTER TABLE `iac_project` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_project` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_project_template`
--

DROP TABLE IF EXISTS `iac_project_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_project_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `template_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__project__template` (`project_id`,`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_project_template`
--

LOCK TABLES `iac_project_template` WRITE;
/*!40000 ALTER TABLE `iac_project_template` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_project_template` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_resource`
--

DROP TABLE IF EXISTS `iac_resource`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_resource` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `res_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `provider` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `module` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `index` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `attrs` json DEFAULT NULL,
  `sensitive_keys` json DEFAULT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_iac_resource_env_id` (`env_id`),
  KEY `idx_iac_resource_task_id` (`task_id`),
  KEY `idx_iac_resource_res_id` (`res_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_resource`
--

LOCK TABLES `iac_resource` WRITE;
/*!40000 ALTER TABLE `iac_resource` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_resource` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_resource_account`
--

DROP TABLE IF EXISTS `iac_resource_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_resource_account` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织ID',
  `name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源账号名称',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源账号描述',
  `params` json DEFAULT NULL COMMENT '账号变量',
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT '资源账号状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org_id__name` (`org_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_resource_account`
--

LOCK TABLES `iac_resource_account` WRITE;
/*!40000 ALTER TABLE `iac_resource_account` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_resource_account` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_resource_drift`
--

DROP TABLE IF EXISTS `iac_resource_drift`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_resource_drift` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `res_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `drift_detail` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_resource_drift`
--

LOCK TABLES `iac_resource_drift` WRITE;
/*!40000 ALTER TABLE `iac_resource_drift` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_resource_drift` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_scan_task`
--

DROP TABLE IF EXISTS `iac_scan_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_scan_task` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pipeline` text COLLATE utf8mb4_unicode_ci,
  `flow` json DEFAULT NULL,
  `curr_step` bigint DEFAULT '0',
  `container_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step_timeout` bigint DEFAULT '3600' COMMENT '执行超时',
  `runner_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('pending','running','approving','rejected','failed','complete','timeout','aborted') COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `message` text COLLATE utf8mb4_unicode_ci,
  `aborting` tinyint(1) DEFAULT NULL,
  `start_at` datetime DEFAULT NULL COMMENT '任务开始时间',
  `end_at` datetime DEFAULT NULL COMMENT '任务结束时间',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_addr` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `revision` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `commit_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `workdir` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `mirror` tinyint(1) DEFAULT NULL,
  `mirror_task_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `policy_status` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `playbook` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tf_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tf_version` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `play_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `variables` json DEFAULT NULL,
  `state_path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `extra_data` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_iac_scan_task_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_scan_task`
--

LOCK TABLES `iac_scan_task` WRITE;
/*!40000 ALTER TABLE `iac_scan_task` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_scan_task` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_storage`
--

DROP TABLE IF EXISTS `iac_storage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_storage` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` mediumblob,
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_storage`
--

LOCK TABLES `iac_storage` WRITE;
/*!40000 ALTER TABLE `iac_storage` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_storage` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_system_cfg`
--

DROP TABLE IF EXISTS `iac_system_cfg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_system_cfg` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '设定名',
  `value` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '设定值',
  `description` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__system_cfg__name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_system_cfg`
--

LOCK TABLES `iac_system_cfg` WRITE;
/*!40000 ALTER TABLE `iac_system_cfg` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_system_cfg` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_task`
--

DROP TABLE IF EXISTS `iac_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_task` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pipeline` text COLLATE utf8mb4_unicode_ci,
  `flow` json DEFAULT NULL,
  `curr_step` bigint DEFAULT '0',
  `container_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step_timeout` bigint DEFAULT '3600' COMMENT '执行超时',
  `runner_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('pending','running','approving','rejected','failed','complete','timeout','aborted') COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `message` text COLLATE utf8mb4_unicode_ci,
  `aborting` tinyint(1) DEFAULT NULL,
  `start_at` datetime DEFAULT NULL COMMENT '任务开始时间',
  `end_at` datetime DEFAULT NULL COMMENT '任务结束时间',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务名称',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_addr` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `revision` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `commit_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `workdir` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `playbook` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tf_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tf_version` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `play_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `targets` json DEFAULT NULL,
  `variables` json DEFAULT NULL,
  `state_path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `extra_data` json DEFAULT NULL,
  `key_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `auto_approve` tinyint(1) DEFAULT '0',
  `stop_on_violation` tinyint(1) DEFAULT '0',
  `result` json DEFAULT NULL,
  `plan_result` json DEFAULT NULL,
  `retry_number` int DEFAULT '0',
  `retry_delay` int DEFAULT '0',
  `retry_able` tinyint(1) DEFAULT '0',
  `callback` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `is_drift_task` tinyint(1) DEFAULT '0',
  `applied` tinyint(1) DEFAULT '0',
  `source` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'manual',
  `source_sys` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_iac_task_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_task`
--

LOCK TABLES `iac_task` WRITE;
/*!40000 ALTER TABLE `iac_task` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_task` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_task_comment`
--

DROP TABLE IF EXISTS `iac_task_comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_task_comment` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `task_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务id',
  `creator` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论人',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论人id',
  `comment` text COLLATE utf8mb4_unicode_ci COMMENT '评论',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_task_comment`
--

LOCK TABLES `iac_task_comment` WRITE;
/*!40000 ALTER TABLE `iac_task_comment` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_task_comment` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_task_step`
--

DROP TABLE IF EXISTS `iac_task_step`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_task_step` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `args` text COLLATE utf8mb4_unicode_ci,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `next_step` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `index` int NOT NULL,
  `status` enum('pending','approving','rejected','running','failed','complete','timeout','aborted') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `exit_code` bigint DEFAULT '0',
  `message` text COLLATE utf8mb4_unicode_ci,
  `start_at` datetime DEFAULT NULL,
  `end_at` datetime DEFAULT NULL,
  `log_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `must_approval` tinyint(1) DEFAULT NULL,
  `approver_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `current_retry_count` int DEFAULT '0',
  `next_retry_time` bigint DEFAULT '0',
  `retry_number` int DEFAULT '0',
  `is_callback` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_task_step`
--

LOCK TABLES `iac_task_step` WRITE;
/*!40000 ALTER TABLE `iac_task_step` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_task_step` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_template`
--

DROP TABLE IF EXISTS `iac_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_template` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称',
  `tpl_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '云模板类型(aliyun，VMware等)',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `vcs_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_full_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_revision` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT 'master',
  `repo_addr` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_token` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `repo_user` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT '状态',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人',
  `workdir` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tf_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `playbook` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `play_vars_file` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `last_scan_task_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `tf_version` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `triggers` text COLLATE utf8mb4_unicode_ci,
  `policy_enable` tinyint(1) DEFAULT '0',
  `key_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org__tpl__name` (`org_id`,`name`,`deleted_at_t`),
  KEY `idx_iac_template_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_template`
--

LOCK TABLES `iac_template` WRITE;
/*!40000 ALTER TABLE `iac_template` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_template` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_token`
--

DROP TABLE IF EXISTS `iac_token`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_token` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `role` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT 'Token状态',
  `expired_at` datetime DEFAULT NULL,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '描述',
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人',
  `env_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `action` enum('apply','plan','destroy') COLLATE utf8mb4_unicode_ci DEFAULT 'plan',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_token`
--

LOCK TABLES `iac_token` WRITE;
/*!40000 ALTER TABLE `iac_token` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_token` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_user`
--

DROP TABLE IF EXISTS `iac_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_user` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at_t` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '姓名',
  `email` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱',
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码',
  `phone` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电话',
  `is_admin` tinyint(1) DEFAULT '0' COMMENT '是否为系统管理员',
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT '用户状态',
  `newbie_guide` json DEFAULT NULL COMMENT '新手引导状态',
  `is_ldap` tinyint(1) DEFAULT '0' COMMENT '是否为ldap账号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__email` (`email`,`deleted_at_t`),
  KEY `idx_iac_user_deleted_at_t` (`deleted_at_t`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_user`
--

LOCK TABLES `iac_user` WRITE;
/*!40000 ALTER TABLE `iac_user` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_user_org`
--

DROP TABLE IF EXISTS `iac_user_org`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_user_org` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户ID',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组织ID',
  `role` enum('admin','complianceManager','member') COLLATE utf8mb4_unicode_ci DEFAULT 'member',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org_id__user_id` (`org_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_user_org`
--

LOCK TABLES `iac_user_org` WRITE;
/*!40000 ALTER TABLE `iac_user_org` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_user_org` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_user_project`
--

DROP TABLE IF EXISTS `iac_user_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_user_project` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户ID',
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `role` enum('manager','approver','operator','guest') COLLATE utf8mb4_unicode_ci DEFAULT 'operator' COMMENT '角色',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__user__project` (`user_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_user_project`
--

LOCK TABLES `iac_user_project` WRITE;
/*!40000 ALTER TABLE `iac_user_project` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_user_project` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_variable`
--

DROP TABLE IF EXISTS `iac_variable`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_variable` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `scope` enum('org','template','project','env') COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` enum('environment','terraform','ansible') COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text COLLATE utf8mb4_unicode_ci,
  `sensitive` tinyint(1) DEFAULT '0',
  `description` text COLLATE utf8mb4_unicode_ci,
  `options` json DEFAULT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `tpl_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `env_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__variable__name` (`org_id`,`project_id`,`tpl_id`,`env_id`,`name`(32),`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_variable`
--

LOCK TABLES `iac_variable` WRITE;
/*!40000 ALTER TABLE `iac_variable` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_variable` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_variable_group`
--

DROP TABLE IF EXISTS `iac_variable_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_variable_group` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` enum('environment','terraform') COLLATE utf8mb4_unicode_ci NOT NULL,
  `creator_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建人',
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `variables` json DEFAULT NULL COMMENT '变量组下的变量',
  `cost_counted` tinyint(1) DEFAULT '0' COMMENT '是否开启费用统计',
  `provider` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源供应平台名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org__variable_group_name` (`org_id`,`name`(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_variable_group`
--

LOCK TABLES `iac_variable_group` WRITE;
/*!40000 ALTER TABLE `iac_variable_group` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_variable_group` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_variable_group_project_rel`
--

DROP TABLE IF EXISTS `iac_variable_group_project_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_variable_group_project_rel` (
  `var_group_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  UNIQUE KEY `idx_var_group_project` (`var_group_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_variable_group_project_rel`
--

LOCK TABLES `iac_variable_group_project_rel` WRITE;
/*!40000 ALTER TABLE `iac_variable_group_project_rel` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_variable_group_project_rel` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_variable_group_rel`
--

DROP TABLE IF EXISTS `iac_variable_group_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_variable_group_rel` (
  `var_group_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `object_type` enum('org','template','project','env') COLLATE utf8mb4_unicode_ci NOT NULL,
  `object_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_variable_group_rel`
--

LOCK TABLES `iac_variable_group_rel` WRITE;
/*!40000 ALTER TABLE `iac_variable_group_rel` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_variable_group_rel` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_vcs`
--

DROP TABLE IF EXISTS `iac_vcs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_vcs` (
  `id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'vcs名称',
  `status` enum('enable','disable') COLLATE utf8mb4_unicode_ci DEFAULT 'enable' COMMENT 'vcs状态',
  `vcs_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'vcs代码库类型',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'vcs代码库地址',
  `vcs_token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '代码库的token值',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique__org_vcs_name` (`org_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_vcs`
--

LOCK TABLES `iac_vcs` WRITE;
/*!40000 ALTER TABLE `iac_vcs` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_vcs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `iac_vcs_pr`
--

DROP TABLE IF EXISTS `iac_vcs_pr`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iac_vcs_pr` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pr_id` bigint DEFAULT NULL,
  `task_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `env_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `vcs_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `iac_vcs_pr`
--

LOCK TABLES `iac_vcs_pr` WRITE;
/*!40000 ALTER TABLE `iac_vcs_pr` DISABLE KEYS */;
/*!40000 ALTER TABLE `iac_vcs_pr` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2022-04-18 16:00:35
