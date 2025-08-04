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
  if (!email || email === 'REDACTED' || email.trim() === '') {
    return '-'
  }
  
  if (email.includes('@')) {
    const username = email.split('@')[0]
    return username === 'REDACTED' ? '-' : username
  }
  
  return email === 'REDACTED' ? '-' : email
}