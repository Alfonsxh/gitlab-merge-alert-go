/**
 * 格式化日期
 */
export function formatDate(dateString: string | null | undefined): string {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

/**
 * 格式化手机号，中间部分用*代替
 */
export function formatPhone(phone: string | null | undefined): string {
  if (!phone || phone.length < 7) return phone || '-'
  const start = phone.substring(0, 3)
  const end = phone.substring(phone.length - 4)
  const middle = '*'.repeat(phone.length - 7)
  return start + middle + end
}

/**
 * 从邮箱中提取用户名
 */
export function extractNameFromEmail(email: string | null | undefined): string {
  if (!email || email.trim() === '') {
    return '-'
  }

  // 使用 includes 判断是否包含 REDACTED
  if (email.includes('REDACTED')) {
    return '-'
  }

  // 如果已经是名字格式（不含@），直接返回
  if (!email.includes('@')) {
    return email
  }

  // 从邮箱提取用户名
  const username = email.split('@')[0]
  // 再次检查提取出的用户名是否包含 REDACTED
  return username.includes('REDACTED') ? '-' : username
}