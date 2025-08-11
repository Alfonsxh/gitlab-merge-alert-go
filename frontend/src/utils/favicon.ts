/**
 * 动态设置页面的 favicon
 * @param iconUrl - 图标 URL，可以是图片地址或 emoji
 */
export function setFavicon(iconUrl: string | null) {
  // 获取或创建 favicon link 元素
  let link = document.querySelector<HTMLLinkElement>('link[rel*="icon"]')
  
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  
  if (!iconUrl) {
    // 如果没有图标，使用默认的
    link.href = '/vite.svg'
    return
  }
  
  // 检查是否是 emoji
  if (iconUrl.length <= 2 && /\p{Emoji}/u.test(iconUrl)) {
    // 将 emoji 转换为 data URL
    const canvas = document.createElement('canvas')
    canvas.width = 32
    canvas.height = 32
    const ctx = canvas.getContext('2d')
    
    if (ctx) {
      ctx.font = '28px serif'
      ctx.textAlign = 'center'
      ctx.textBaseline = 'middle'
      ctx.fillText(iconUrl, 16, 16)
      
      link.href = canvas.toDataURL()
    }
  } else {
    // 直接使用图片 URL
    link.href = iconUrl
  }
}

/**
 * 从用户头像生成 favicon
 * @param avatarUrl - 头像 URL
 */
export function setFaviconFromAvatar(avatarUrl: string | null) {
  if (!avatarUrl) {
    setFavicon(null)
    return
  }
  
  // 如果是 emoji，直接使用
  if (avatarUrl.length <= 2 && /\p{Emoji}/u.test(avatarUrl)) {
    setFavicon(avatarUrl)
    return
  }
  
  // 如果是图片 URL，创建一个小尺寸的版本
  const img = new Image()
  img.crossOrigin = 'anonymous'
  
  img.onload = () => {
    const canvas = document.createElement('canvas')
    canvas.width = 32
    canvas.height = 32
    const ctx = canvas.getContext('2d')
    
    if (ctx) {
      // 绘制圆形头像
      ctx.beginPath()
      ctx.arc(16, 16, 16, 0, Math.PI * 2)
      ctx.closePath()
      ctx.clip()
      
      ctx.drawImage(img, 0, 0, 32, 32)
      
      setFavicon(canvas.toDataURL())
    }
  }
  
  img.onerror = () => {
    // 加载失败，使用默认
    setFavicon(null)
  }
  
  img.src = avatarUrl
}