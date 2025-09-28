#!/usr/bin/env node

/**
 * 图标生成脚本
 * 用于从 SVG 文件生成不同尺寸的 PNG 图标
 * 
 * 使用方法：
 * 1. 安装依赖：npm install canvas
 * 2. 运行脚本：node generate-icons.js
 */

const fs = require('fs');
const path = require('path');

// 需要生成的图标尺寸
const ICON_SIZES = [72, 96, 128, 144, 152, 192, 384, 512];

// 基础图标配置
const ICON_CONFIG = {
  background: '#667eea',
  foreground: '#ffffff',
  padding: 0.1 // 10% padding
};

/**
 * 创建一个简单的PNG图标（使用Canvas API）
 */
function createIcon(size) {
  try {
    const { createCanvas } = require('canvas');
    
    const canvas = createCanvas(size, size);
    const ctx = canvas.getContext('2d');
    
    // 绘制背景
    const gradient = ctx.createLinearGradient(0, 0, size, size);
    gradient.addColorStop(0, '#667eea');
    gradient.addColorStop(1, '#764ba2');
    
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, size, size);
    
    // 圆角处理（简化版）
    ctx.globalCompositeOperation = 'destination-in';
    ctx.beginPath();
    const radius = size * 0.15;
    ctx.roundRect(0, 0, size, size, radius);
    ctx.fill();
    
    // 重置合成模式
    ctx.globalCompositeOperation = 'source-over';
    
    // 绘制电脑图标
    const padding = size * ICON_CONFIG.padding;
    const innerSize = size - padding * 2;
    const screenWidth = innerSize * 0.5;
    const screenHeight = innerSize * 0.35;
    const screenX = (size - screenWidth) / 2;
    const screenY = (size - screenHeight) / 2;
    
    // 屏幕外框
    ctx.fillStyle = ICON_CONFIG.foreground;
    ctx.globalAlpha = 0.95;
    ctx.fillRect(screenX, screenY, screenWidth, screenHeight);
    
    // 屏幕内容
    ctx.fillStyle = '#667eea';
    const innerPadding = size * 0.02;
    ctx.fillRect(
      screenX + innerPadding, 
      screenY + innerPadding, 
      screenWidth - innerPadding * 2, 
      screenHeight - innerPadding * 2
    );
    
    // 屏幕内容线条
    ctx.strokeStyle = ICON_CONFIG.foreground;
    ctx.globalAlpha = 0.8;
    ctx.lineWidth = Math.max(1, size * 0.006);
    
    const lineY1 = screenY + screenHeight * 0.3;
    const lineY2 = screenY + screenHeight * 0.5;
    const lineY3 = screenY + screenHeight * 0.7;
    
    ctx.beginPath();
    ctx.moveTo(screenX + innerPadding * 2, lineY1);
    ctx.lineTo(screenX + screenWidth * 0.7, lineY1);
    ctx.stroke();
    
    ctx.beginPath();
    ctx.moveTo(screenX + innerPadding * 2, lineY2);
    ctx.lineTo(screenX + screenWidth * 0.9, lineY2);
    ctx.stroke();
    
    ctx.beginPath();
    ctx.moveTo(screenX + innerPadding * 2, lineY3);
    ctx.lineTo(screenX + screenWidth * 0.6, lineY3);
    ctx.stroke();
    
    // 底座
    ctx.globalAlpha = 0.9;
    ctx.fillStyle = ICON_CONFIG.foreground;
    const standWidth = screenWidth * 0.3;
    const standHeight = size * 0.04;
    const standX = (size - standWidth) / 2;
    const standY = screenY + screenHeight + size * 0.02;
    
    ctx.fillRect(standX, standY, standWidth, standHeight);
    
    // 底座基座
    const baseWidth = screenWidth * 0.6;
    const baseHeight = size * 0.02;
    const baseX = (size - baseWidth) / 2;
    const baseY = standY + standHeight;
    
    ctx.fillRect(baseX, baseY, baseWidth, baseHeight);
    
    return canvas.toBuffer('image/png');
    
  } catch (error) {
    console.error('无法使用Canvas生成图标，请手动创建图标文件');
    return null;
  }
}

/**
 * 生成所有尺寸的图标
 */
function generateIcons() {
  const iconsDir = path.join(__dirname, 'static', 'icons');
  
  // 确保图标目录存在
  if (!fs.existsSync(iconsDir)) {
    fs.mkdirSync(iconsDir, { recursive: true });
  }
  
  console.log('开始生成图标文件...');
  
  for (const size of ICON_SIZES) {
    const iconBuffer = createIcon(size);
    
    if (iconBuffer) {
      const filename = `icon-${size}x${size}.png`;
      const filepath = path.join(iconsDir, filename);
      
      fs.writeFileSync(filepath, iconBuffer);
      console.log(`✓ 生成 ${filename}`);
    } else {
      console.log(`✗ 无法生成 icon-${size}x${size}.png`);
    }
  }
  
  console.log('\n图标生成完成！');
  console.log('如果生成失败，请：');
  console.log('1. 安装 canvas 依赖：npm install canvas');
  console.log('2. 或手动创建以下尺寸的图标文件：');
  ICON_SIZES.forEach(size => {
    console.log(`   - static/icons/icon-${size}x${size}.png`);
  });
}

/**
 * 创建快捷图标（简化版）
 */
function createShortcutIcons() {
  const iconsDir = path.join(__dirname, 'static', 'icons');
  
  // 为快捷方式创建简单的图标文件
  const shortcuts = [
    { name: 'network-shortcut.png', size: 96, emoji: '🌐' },
    { name: 'scan-shortcut.png', size: 96, emoji: '🔍' }
  ];
  
  shortcuts.forEach(shortcut => {
    try {
      const { createCanvas } = require('canvas');
      
      const canvas = createCanvas(shortcut.size, shortcut.size);
      const ctx = canvas.getContext('2d');
      
      // 简单的背景
      ctx.fillStyle = '#667eea';
      ctx.fillRect(0, 0, shortcut.size, shortcut.size);
      
      // 使用emoji作为图标（简化版）
      ctx.fillStyle = '#ffffff';
      ctx.font = `${shortcut.size * 0.6}px Arial`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(shortcut.emoji, shortcut.size / 2, shortcut.size / 2);
      
      const buffer = canvas.toBuffer('image/png');
      fs.writeFileSync(path.join(iconsDir, shortcut.name), buffer);
      console.log(`✓ 生成 ${shortcut.name}`);
      
    } catch (error) {
      console.log(`✗ 无法生成 ${shortcut.name}`);
    }
  });
}

// 主函数
if (require.main === module) {
  generateIcons();
  createShortcutIcons();
}

module.exports = { generateIcons, createIcon };


