#!/usr/bin/env python3
"""
AI背景移除服务
支持CPU和GPU处理，使用rembg库进行智能抠图
"""

import sys
import json
import base64
import io
import os
from PIL import Image, ImageEnhance
import numpy as np

# 尝试导入GPU加速库
try:
    import torch
    import torchvision
    GPU_AVAILABLE = torch.cuda.is_available()
    print(f"GPU Available: {GPU_AVAILABLE}", file=sys.stderr)
except ImportError:
    GPU_AVAILABLE = False
    print("PyTorch not available, using CPU only", file=sys.stderr)

# 导入rembg
try:
    from rembg import remove, new_session
    REMBG_AVAILABLE = True
    print("rembg loaded successfully", file=sys.stderr)
except ImportError:
    REMBG_AVAILABLE = False
    print("rembg not available, using fallback", file=sys.stderr)

def setup_device():
    """设置计算设备"""
    if GPU_AVAILABLE:
        device = 'cuda'
        print(f"Using GPU: {torch.cuda.get_device_name()}", file=sys.stderr)
    else:
        device = 'cpu'
        print("Using CPU", file=sys.stderr)
    return device

def remove_background_ai(image_data, use_gpu=True):
    """使用AI模型移除背景"""
    try:
        # 解码base64图片
        if image_data.startswith('data:image/'):
            image_data = image_data.split(',')[1]
        
        image_bytes = base64.b64decode(image_data)
        input_image = Image.open(io.BytesIO(image_bytes)).convert('RGB')
        
        print(f"Input image size: {input_image.size}", file=sys.stderr)
        
        if REMBG_AVAILABLE:
            # 使用rembg进行背景移除
            device = setup_device() if use_gpu else 'cpu'
            
            # 选择最适合人像的模型
            if GPU_AVAILABLE and use_gpu:
                session = new_session('u2net_human_seg')  # 专门针对人像优化
            else:
                session = new_session('u2net')  # 通用模型，CPU友好
            
            print(f"Using model: {session.model_name}", file=sys.stderr)
            
            # 移除背景
            output_image = remove(input_image, session=session)
            
            # 转换为RGBA
            if output_image.mode != 'RGBA':
                output_image = output_image.convert('RGBA')
                
            print(f"Output image size: {output_image.size}, mode: {output_image.mode}", file=sys.stderr)
            
        else:
            # 回退到简单的边缘检测算法
            print("Using fallback edge detection", file=sys.stderr)
            output_image = remove_background_fallback(input_image)
        
        # 后处理：增强边缘
        output_image = enhance_edges(output_image)
        
        return output_image
        
    except Exception as e:
        print(f"AI background removal failed: {str(e)}", file=sys.stderr)
        # 回退处理
        return remove_background_fallback(input_image)

def enhance_edges(image):
    """增强边缘效果"""
    try:
        if image.mode != 'RGBA':
            image = image.convert('RGBA')
        
        # 分离alpha通道
        r, g, b, a = image.split()
        
        # 对alpha通道进行轻微的锐化
        alpha_enhancer = ImageEnhance.Sharpness(a)
        enhanced_alpha = alpha_enhancer.enhance(1.2)
        
        # 重新组合
        enhanced_image = Image.merge('RGBA', (r, g, b, enhanced_alpha))
        
        return enhanced_image
    except Exception as e:
        print(f"Edge enhancement failed: {str(e)}", file=sys.stderr)
        return image

def remove_background_fallback(input_image):
    """回退算法：简单的边缘检测"""
    try:
        import cv2
        from skimage import filters, morphology, measure
        
        # 转换为numpy数组
        img_array = np.array(input_image)
        
        # 转换为灰度
        gray = cv2.cvtColor(img_array, cv2.COLOR_RGB2GRAY)
        
        # 使用Otsu阈值
        threshold = filters.threshold_otsu(gray)
        binary = gray > threshold
        
        # 形态学操作
        binary = morphology.closing(binary, morphology.disk(3))
        binary = morphology.opening(binary, morphology.disk(2))
        
        # 找到最大连通区域（假设为前景）
        labeled = measure.label(binary)
        regions = measure.regionprops(labeled)
        
        if regions:
            # 选择面积最大的区域
            largest_region = max(regions, key=lambda x: x.area)
            mask = labeled == largest_region.label
        else:
            mask = binary
        
        # 创建RGBA图像
        output_array = np.zeros((*img_array.shape[:2], 4), dtype=np.uint8)
        output_array[:, :, :3] = img_array
        output_array[:, :, 3] = (mask * 255).astype(np.uint8)
        
        return Image.fromarray(output_array, 'RGBA')
        
    except Exception as e:
        print(f"Fallback algorithm failed: {str(e)}", file=sys.stderr)
        # 最终回退：返回原图
        rgba_image = input_image.convert('RGBA')
        return rgba_image

def replace_background(foreground_image, background_color, target_width, target_height):
    """替换背景颜色并调整尺寸"""
    try:
        # 解析背景颜色
        if background_color.startswith('#'):
            background_color = background_color[1:]
        
        r = int(background_color[0:2], 16)
        g = int(background_color[2:4], 16) 
        b = int(background_color[4:6], 16)
        bg_color = (r, g, b, 255)
        
        # 创建目标尺寸的背景
        background = Image.new('RGBA', (target_width, target_height), bg_color)
        
        # 计算前景图像的缩放比例
        scale = min(target_width / foreground_image.width, target_height / foreground_image.height)
        new_width = int(foreground_image.width * scale)
        new_height = int(foreground_image.height * scale)
        
        # 缩放前景图像
        foreground_resized = foreground_image.resize((new_width, new_height), Image.Resampling.LANCZOS)
        
        # 计算居中位置
        x_offset = (target_width - new_width) // 2
        y_offset = (target_height - new_height) // 2
        
        # 合成图像
        background.paste(foreground_resized, (x_offset, y_offset), foreground_resized)
        
        return background
        
    except Exception as e:
        print(f"Background replacement failed: {str(e)}", file=sys.stderr)
        raise

def process_id_photo(request_data):
    """处理证件照请求"""
    try:
        image_data = request_data['imageData']
        background_color = request_data.get('backgroundColor', '#FFFFFF')
        width = int(request_data.get('width', 295))
        height = int(request_data.get('height', 413))
        quality = float(request_data.get('quality', 0.8))
        output_format = request_data.get('format', 'jpeg')
        use_gpu = request_data.get('useGPU', GPU_AVAILABLE)
        
        print(f"Processing: {width}x{height}, bg:{background_color}, quality:{quality}, GPU:{use_gpu}", file=sys.stderr)
        
        # 移除背景
        foreground_image = remove_background_ai(image_data, use_gpu)
        
        # 替换背景
        final_image = replace_background(foreground_image, background_color, width, height)
        
        # 转换输出格式
        if output_format.lower() == 'jpeg':
            final_image = final_image.convert('RGB')
        
        # 保存到内存
        output_buffer = io.BytesIO()
        if output_format.lower() == 'jpeg':
            final_image.save(output_buffer, format='JPEG', quality=int(quality * 100), optimize=True)
        else:
            final_image.save(output_buffer, format='PNG', optimize=True)
        
        output_buffer.seek(0)
        
        # 编码为base64
        encoded_image = base64.b64encode(output_buffer.read()).decode('utf-8')
        data_url = f"data:image/{output_format};base64,{encoded_image}"
        
        file_size = len(output_buffer.getvalue())
        
        return {
            'success': True,
            'imageData': data_url,
            'fileSize': file_size,
            'width': width,
            'height': height,
            'usedGPU': use_gpu and GPU_AVAILABLE
        }
        
    except Exception as e:
        print(f"Process failed: {str(e)}", file=sys.stderr)
        return {
            'success': False,
            'error': str(e),
            'usedGPU': False
        }

def main():
    """主函数"""
    try:
        # 读取JSON输入
        input_data = sys.stdin.read().strip()
        if not input_data:
            raise ValueError("No input data")
        
        request_data = json.loads(input_data)
        
        # 处理请求
        result = process_id_photo(request_data)
        
        # 输出结果
        print(json.dumps(result))
        
    except Exception as e:
        print(json.dumps({
            'success': False,
            'error': str(e),
            'usedGPU': False
        }))

if __name__ == "__main__":
    main()
