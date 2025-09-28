#!/usr/bin/env python3
"""
快速创建PWA所需的图标文件
"""
from PIL import Image, ImageDraw, ImageFont
import os

# 创建图标目录
icons_dir = "static/icons"
if not os.path.exists(icons_dir):
    os.makedirs(icons_dir)

# 需要的图标尺寸
sizes = [72, 96, 128, 144, 152, 192, 384, 512]

# 颜色配置
bg_color = "#667eea"
text_color = "white"

def create_icon(size):
    # 创建图像
    img = Image.new('RGB', (size, size), bg_color)
    draw = ImageDraw.Draw(img)
    
    # 绘制简单的电脑图标
    # 屏幕
    screen_size = size * 0.6
    screen_x = (size - screen_size) // 2
    screen_y = (size - screen_size) // 2
    
    draw.rectangle([screen_x, screen_y, screen_x + screen_size, screen_y + screen_size * 0.7], fill="white")
    
    # 底座
    stand_width = screen_size * 0.3
    stand_x = (size - stand_width) // 2
    stand_y = screen_y + screen_size * 0.7
    draw.rectangle([stand_x, stand_y, stand_x + stand_width, stand_y + size * 0.1], fill="white")
    
    # 底座基座
    base_width = screen_size * 0.6
    base_x = (size - base_width) // 2
    base_y = stand_y + size * 0.1
    draw.rectangle([base_x, base_y, base_x + base_width, base_y + size * 0.05], fill="white")
    
    return img

def main():
    print("开始生成PWA图标...")
    
    for size in sizes:
        try:
            icon = create_icon(size)
            filename = f"{icons_dir}/icon-{size}x{size}.png"
            icon.save(filename)
            print(f"✅ 已生成 {filename}")
        except Exception as e:
            print(f"❌ 生成 {size}x{size} 图标失败: {e}")
    
    print("\n🎉 图标生成完成!")
    print("PWA安装功能现在应该可以正常工作了")

if __name__ == "__main__":
    main()


