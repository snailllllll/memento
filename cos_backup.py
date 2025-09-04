#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
腾讯云COS备份脚本
将本地文件夹备份到腾讯云对象存储
使用方法: python cos_backup.py <本地文件夹路径>
"""

import os
import sys
import argparse
from typing import List, Set
from qcloud_cos import CosConfig
from qcloud_cos import CosS3Client
from dotenv import load_dotenv
import logging

# 加载环境变量
load_dotenv()

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class COSBackup:
    def __init__(self):
        """初始化COS客户端"""
        self.secret_id = os.getenv('COS_SECRET_ID')
        self.secret_key = os.getenv('COS_SECRET_KEY')
        self.region = os.getenv('COS_REGION', 'ap-beijing')
        self.bucket = os.getenv('COS_BUCKET')
        
        if not all([self.secret_id, self.secret_key, self.bucket]):
            raise ValueError("请设置环境变量: COS_SECRET_ID, COS_SECRET_KEY, COS_BUCKET")
        
        config = CosConfig(
            Region=self.region,
            SecretId=self.secret_id,
            SecretKey=self.secret_key
        )
        self.client = CosS3Client(config)
        
    def list_cos_files(self, cos_prefix: str) -> Set[str]:
        """获取COS指定前缀下的所有文件列表"""
        files = set()
        marker = ""
        
        while True:
            response = self.client.list_objects(
                Bucket=self.bucket,
                Prefix=cos_prefix,
                Marker=marker,
                MaxKeys=1000
            )
            
            if 'Contents' in response:
                for content in response['Contents']:
                    # 移除前缀，只保留相对路径
                    file_path = content['Key']
                    if file_path.startswith(cos_prefix):
                        file_path = file_path[len(cos_prefix):].lstrip('/')
                    files.add(file_path)
            
            if response.get('IsTruncated') == 'true':
                marker = response['NextMarker']
            else:
                break
                
        return files
    
    def list_local_files(self, local_dir: str) -> Set[str]:
        """获取本地文件夹中的所有文件列表"""
        files = set()
        local_dir = os.path.abspath(local_dir)
        
        if not os.path.exists(local_dir):
            raise FileNotFoundError(f"本地目录不存在: {local_dir}")
        
        for root, dirs, filenames in os.walk(local_dir):
            for filename in filenames:
                full_path = os.path.join(root, filename)
                # 计算相对路径
                rel_path = os.path.relpath(full_path, local_dir)
                files.add(rel_path.replace(os.sep, '/'))
        
        return files
    
    def upload_file(self, local_path: str, cos_key: str) -> bool:
        """上传单个文件到COS"""
        try:
            logger.info(f"正在上传: {local_path} -> {cos_key}")
            self.client.upload_file(
                Bucket=self.bucket,
                LocalFilePath=local_path,
                Key=cos_key
            )
            logger.info(f"上传成功: {cos_key}")
            return True
        except Exception as e:
            logger.error(f"上传失败: {local_path} -> {cos_key}, 错误: {str(e)}")
            return False
    
    def backup_directory(self, local_dir: str, cos_prefix: str = None) -> None:
        """备份本地目录到COS"""
        local_dir = os.path.abspath(local_dir)
        
        if cos_prefix is None:
            cos_prefix = os.path.basename(local_dir.rstrip('/'))
        
        logger.info(f"开始备份目录: {local_dir} -> COS前缀: {cos_prefix}")
        
        try:
            # 获取本地文件列表
            local_files = self.list_local_files(local_dir)
            logger.info(f"本地文件数量: {len(local_files)}")
            
            if not local_files:
                logger.warning("本地目录为空，无需备份")
                return
            
            # 获取COS文件列表
            cos_files = self.list_cos_files(cos_prefix)
            logger.info(f"COS文件数量: {len(cos_files)}")
            
            # 计算需要上传的文件（本地有但COS没有）
            files_to_upload = local_files - cos_files
            logger.info(f"需要上传的文件数量: {len(files_to_upload)}")
            
            if not files_to_upload:
                logger.info("所有文件已存在于COS，无需上传")
                return
            
            # 上传文件
            success_count = 0
            for file_path in files_to_upload:
                local_file_path = os.path.join(local_dir, file_path)
                cos_key = f"{cos_prefix}/{file_path}".replace('//', '/')
                
                if self.upload_file(local_file_path, cos_key):
                    success_count += 1
            
            logger.info(f"备份完成！成功上传: {success_count}/{len(files_to_upload)} 个文件")
            
        except Exception as e:
            logger.error(f"备份过程中发生错误: {str(e)}")
            raise

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='腾讯云COS备份脚本')
    parser.add_argument('local_dir', help='本地目录路径')
    parser.add_argument('--prefix', help='COS前缀（默认为本地目录名）', default=None)
    parser.add_argument('--verbose', '-v', action='store_true', help='显示详细日志')
    
    args = parser.parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    try:
        backup = COSBackup()
        backup.backup_directory(args.local_dir, args.prefix)
    except Exception as e:
        logger.error(f"备份失败: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main()