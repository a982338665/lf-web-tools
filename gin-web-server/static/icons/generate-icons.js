// 在浏览器控制台运行此脚本来快速生成图标
const sizes = [72, 96, 128, 144, 152, 192, 384, 512];

function createAndDownloadIcon(size) {
    const canvas = document.createElement('canvas');
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext('2d');
    
    // 渐变背景
    const gradient = ctx.createLinearGradient(0, 0, size, size);
    gradient.addColorStop(0, '#667eea');
    gradient.addColorStop(1, '#764ba2');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, size, size);
    
    // 电脑图标
    ctx.fillStyle = 'white';
    ctx.globalAlpha = 0.9;
    const screenSize = size * 0.5;
    const x = (size - screenSize) / 2;
    const y = (size - screenSize) / 2;
    ctx.fillRect(x, y, screenSize, screenSize * 0.6);
    
    // 底座
    ctx.fillRect(x + screenSize * 0.3, y + screenSize * 0.6, screenSize * 0.4, size * 0.08);
    ctx.fillRect(x + screenSize * 0.1, y + screenSize * 0.68, screenSize * 0.8, size * 0.04);
    
    // 下载
    canvas.toBlob(blob => {
        const link = document.createElement('a');
        link.download = `icon-${size}x${size}.png`;
        link.href = URL.createObjectURL(blob);
        link.click();
    });
}

// 生成所有图标
sizes.forEach((size, index) => {
    setTimeout(() => createAndDownloadIcon(size), index * 200);
});

console.log('图标生成完成，请将下载的PNG文件放入 static/icons/ 目录');


