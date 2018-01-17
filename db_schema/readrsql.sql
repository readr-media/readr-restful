-- MySQL dump 10.13  Distrib 5.7.20, for osx10.13 (x86_64)
--
-- Host: 127.0.0.1    Database: memberdb
-- ------------------------------------------------------
-- Server version	5.7.14-google-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `memberdb`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `memberdb` /*!40100 DEFAULT CHARACTER SET utf8 */;

USE `memberdb`;

--
-- Table structure for table `article_comments`
--

DROP TABLE IF EXISTS `article_comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `article_comments` (
  `post_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `comment_id` varchar(24) NOT NULL,
  PRIMARY KEY (`post_id`,`comment_id`),
  UNIQUE KEY `post_id` (`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comment_infos`
--

DROP TABLE IF EXISTS `comment_infos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `comment_infos` (
  `comment_id` varchar(24) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `like_amount` int(11) DEFAULT NULL,
  `comment_amount` int(11) DEFAULT NULL,
  `reply_comment` varchar(24) DEFAULT NULL,
  `comment_text` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `following_members`
--

DROP TABLE IF EXISTS `following_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `following_members` (
  `member_id` varchar(48) NOT NULL,
  `custom_editor` varchar(48) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`member_id`,`custom_editor`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `following_posts`
--

DROP TABLE IF EXISTS `following_posts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `following_posts` (
  `member_id` varchar(48) NOT NULL,
  `post_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`member_id`,`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `following_projects`
--

DROP TABLE IF EXISTS `following_projects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `following_projects` (
  `member_id` varchar(48) NOT NULL,
  `project_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`member_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `members`
--

DROP TABLE IF EXISTS `members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `members` (
  `member_id` varchar(48) NOT NULL,
  `mail` varchar(48) DEFAULT NULL,
  `register_mode` varchar(36) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `password` binary(64) DEFAULT NULL,
  `salt` binary(32) DEFAULT NULL,
  `social_id` varchar(128) DEFAULT NULL,
  `nickname` varchar(128) DEFAULT NULL,
  `description` text,
  `custom_editor` tinyint(1) DEFAULT '0',
  `role` int(11) DEFAULT '1',
  `profile_image` varchar(256) DEFAULT NULL,
  `hide_profile` tinyint(1) DEFAULT '0',
  `profile_push` tinyint(1) DEFAULT '1',
  `post_push` tinyint(1) DEFAULT '1',
  `comment_push` tinyint(1) DEFAULT '1',
  `active` int(11) NOT NULL DEFAULT '1',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` varchar(48) DEFAULT NULL,
  `birthday` date DEFAULT NULL,
  `gender` char(1) DEFAULT NULL,
  `work` varchar(48) DEFAULT NULL,
  `name` varchar(128) DEFAULT NULL,
  `identity` varchar(24) DEFAULT NULL,
  PRIMARY KEY (`member_id`),
  KEY `mail_nick_ceditor` (`mail`,`nickname`,`custom_editor`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `permissions`
--

DROP TABLE IF EXISTS `permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `permissions` (
  `role` int(11) NOT NULL,
  `object` varchar(50) NOT NULL,
  `permission` int(11) DEFAULT '1',
  PRIMARY KEY (`role`,`object`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `posts`
--

DROP TABLE IF EXISTS `posts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `posts` (
  `post_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `like_amount` int(11) DEFAULT NULL,
  `comment_amount` int(11) DEFAULT NULL,
  `title` varchar(256) DEFAULT NULL,
  `content` text,
  `link` varchar(128) DEFAULT NULL,
  `author` varchar(48) DEFAULT NULL,
  `og_title` varchar(256) DEFAULT NULL,
  `og_description` varchar(256) DEFAULT NULL,
  `og_image` varchar(128) DEFAULT NULL,
  `active` int(11) DEFAULT '1',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` varchar(48) DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  `link_title` varchar(256) DEFAULT NULL,
  `link_description` varchar(256) DEFAULT NULL,
  `link_image` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`post_id`),
  UNIQUE KEY `post_id` (`post_id`),
  KEY `title_link` (`title`,`link`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `project_comments`
--

DROP TABLE IF EXISTS `project_comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `project_comments` (
  `project_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `comment_id` varchar(24) NOT NULL,
  PRIMARY KEY (`project_id`,`comment_id`),
  UNIQUE KEY `project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `projects`
--

DROP TABLE IF EXISTS `projects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `projects` (
  `project_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `like_amount` int(11) DEFAULT NULL,
  `comment_amount` int(11) DEFAULT NULL,
  `hero_image` varchar(256) DEFAULT NULL,
  `title` varchar(256) DEFAULT NULL,
  `description` text,
  `author` text,
  `post_id` bigint(20) unsigned NOT NULL,
  `og_title` varchar(256) DEFAULT NULL,
  `og_description` varchar(256) DEFAULT NULL,
  `og_image` varchar(128) DEFAULT NULL,
  `active` int(11) DEFAULT NULL,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` varchar(48) DEFAULT NULL,
  `published_at` datetime DEFAULT NULL,
  PRIMARY KEY (`project_id`),
  UNIQUE KEY `project_id` (`project_id`),
  KEY `title_postid` (`title`,`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `roles` (
  `role` int(11) NOT NULL,
  `name` varchar(12) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `testdrop`
--

DROP TABLE IF EXISTS `testdrop`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `testdrop` (
  `id` varchar(15) NOT NULL,
  `name` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2018-01-17 10:07:29
