#!/bin/bash

# 测试认证功能的脚本

BASE_URL="http://localhost:1688/api/v1"

echo "=== 测试认证功能 ==="
echo ""

# 1. 测试登录（使用默认管理员账户）
echo "1. 测试登录..."
RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }')

echo "响应: $RESPONSE"
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "登录失败！"
  exit 1
fi

echo "登录成功！获取到 Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# 2. 测试获取用户信息
echo "2. 测试获取当前用户信息..."
curl -s -X GET "${BASE_URL}/auth/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# 3. 测试获取账户列表（需要管理员权限）
echo "3. 测试获取账户列表（管理员权限）..."
curl -s -X GET "${BASE_URL}/accounts" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# 4. 测试创建普通用户账户
echo "4. 测试创建普通用户账户..."
CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/accounts" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "email": "testuser@example.com",
    "role": "user"
  }')

echo "响应: $CREATE_RESPONSE"
echo ""

# 5. 测试普通用户登录
echo "5. 测试普通用户登录..."
USER_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

USER_TOKEN=$(echo "$USER_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$USER_TOKEN" ]; then
  echo "普通用户登录成功！"
  
  # 6. 测试普通用户访问账户列表（应该被拒绝）
  echo ""
  echo "6. 测试普通用户访问账户列表（应该被拒绝）..."
  curl -s -X GET "${BASE_URL}/accounts" \
    -H "Authorization: Bearer $USER_TOKEN"
  echo ""
fi

# 7. 测试数据隔离 - 管理员创建项目
echo ""
echo "7. 测试数据隔离 - 管理员创建项目..."
curl -s -X POST "${BASE_URL}/projects" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Project",
    "gitlab_project_id": 12345,
    "gitlab_url": "https://gitlab.com/admin/project"
  }' | jq .

# 8. 测试数据隔离 - 普通用户创建项目
echo ""
echo "8. 测试数据隔离 - 普通用户创建项目..."
curl -s -X POST "${BASE_URL}/projects" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User Project",
    "gitlab_project_id": 67890,
    "gitlab_url": "https://gitlab.com/user/project"
  }' | jq .

# 9. 测试普通用户查看项目（只能看到自己的）
echo ""
echo "9. 测试普通用户查看项目列表（只能看到自己的）..."
curl -s -X GET "${BASE_URL}/projects" \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# 10. 测试管理员查看项目（能看到所有的）
echo ""
echo "10. 测试管理员查看项目列表（能看到所有的）..."
curl -s -X GET "${BASE_URL}/projects" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

echo ""
echo "=== 测试完成 ==="