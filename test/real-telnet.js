const net = require('net');

/**
 * 真正的Telnet端口检测 - 只能在Node.js环境中使用
 * 这就是浏览器中无法实现的功能
 */

// 真正的TCP连接函数
function telnetPort(host, port, timeout = 3000) {
    return new Promise((resolve) => {
        // console.log(`🔍 Telnet连接测试: ${host}:${port}`);

        const socket = new net.Socket();
        let isResolved = false;

        // 设置超时
        const timer = setTimeout(() => {
            if (!isResolved) {
                isResolved = true;
                socket.destroy();
                // console.log(`⏰ 端口 ${port}: 连接超时`);
                resolve({ status: 'timeout', port, host });
            }
        }, timeout);

        // 连接成功
        socket.on('connect', () => {
            if (!isResolved) {
                isResolved = true;
                clearTimeout(timer);
                // console.log(`✅ 端口 ${port}: 连接成功!`);

                // 尝试发送telnet协商数据
                try {
                    socket.write('\xFF\xFE\x01'); // IAC DONT ECHO
                    socket.write('hello\r\n');
                } catch (e) {
                    // console.log(`⚠️ 端口 ${port}: 无法发送数据，但连接成功`);
                }

                socket.end();
                resolve({ status: 'open', port, host });
            }
        });

        // 连接失败
        socket.on('error', (error) => {
            if (!isResolved) {
                isResolved = true;
                clearTimeout(timer);

                if (error.code === 'ECONNREFUSED') {
                    // console.log(`❌ 端口 ${port}: 连接被拒绝（端口关闭）`);
                    resolve({ status: 'closed', port, host });
                } else if (error.code === 'ENOTFOUND') {
                    // console.log(`❌ 端口 ${port}: 主机不存在`);
                    resolve({ status: 'host_not_found', port, host });
                } else {
                    // console.log(`❌ 端口 ${port}: ${error.message}`);
                    resolve({ status: 'error', port, host, error: error.message });
                }
            }
        });

        // 连接关闭
        socket.on('close', () => {
            if (!isResolved) {
                isResolved = true;
                clearTimeout(timer);
                // console.log(`🔌 端口 ${port}: 连接已关闭`);
                resolve({ status: 'closed', port, host });
            }
        });

        // 开始连接
        try {
            socket.connect(port, host);
        } catch (error) {
            if (!isResolved) {
                isResolved = true;
                clearTimeout(timer);
                // console.log(`❌ 端口 ${port}: 连接异常 - ${error.message}`);
                resolve({ status: 'error', port, host, error: error.message });
            }
        }
    });
}

// 批量测试端口
async function batchTelnetTest(host = 'localhost', ports = []) {
    console.log(`\n🚀 开始批量Telnet测试 ${host}...`);
    console.log(`📋 测试端口: ${ports.join(', ')}`);
    console.log('=====================================\n');

    const results = [];

    // 并行测试所有端口（真正的并发）
    const promises = ports.map(port => telnetPort(host, port));
    const testResults = await Promise.all(promises);

    // 汇总结果
    const summary = {
        open: [],
        closed: [],
        timeout: [],
        error: []
    };

    testResults.forEach(result => {
        results.push(result);
        if (summary[result.status]) {
            summary[result.status].push(result.port);
        }
    });

    // 输出结果
    console.log('\n📊 测试结果汇总:');
    console.log('=====================================');

    if (summary.open.length > 0) {
        console.log(`🟢 开放端口 (${summary.open.length}): ${summary.open.join(', ')}`);
    }

    if (summary.closed.length > 0) {
        console.log(`🔴 关闭端口 (${summary.closed.length}): ${summary.closed.join(', ')}`);
    }

    if (summary.timeout.length > 0) {
        console.log(`⏰ 超时端口 (${summary.timeout.length}): ${summary.timeout.join(', ')}`);
    }

    if (summary.error.length > 0) {
        console.log(`❌ 错误端口 (${summary.error.length}): ${summary.error.join(', ')}`);
    }

    console.log(`\n🎯 总计测试: ${ports.length} 个端口`);
    console.log(`⚡ 测试完成时间: ${new Date().toLocaleString()}\n`);

    return results;
}
// 分批并发处理函数
async function batchProcessPorts(ports, host, batchSize = 100) {
    const results = [];
    const totalPorts = ports.length;
    let processedCount = 0;

    console.log(`📦 使用分批处理模式，每批 ${batchSize} 个端口`);

    // 分批处理
    for (let i = 0; i < ports.length; i += batchSize) {
        const batch = ports.slice(i, i + batchSize);
        const batchNumber = Math.floor(i / batchSize) + 1;
        const totalBatches = Math.ceil(totalPorts / batchSize);

        // 显示当前批次进度（单行覆盖）
        const progress = ((processedCount / totalPorts) * 100).toFixed(1);
        process.stdout.write(`\r⏳ 批次 ${batchNumber}/${totalBatches} | 端口 ${batch[0]}-${batch[batch.length-1]} | 进度 ${progress}% (${processedCount}/${totalPorts})`);

        try {
            const batchPromises = batch.map(port => telnetPort(host, port, 2000)); // 减少超时时间
            const batchResults = await Promise.all(batchPromises);

            results.push(...batchResults);
            processedCount += batch.length;

            // 更新进度条
            const newProgress = ((processedCount / totalPorts) * 100).toFixed(1);
            process.stdout.write(`\r⚡ 批次 ${batchNumber}/${totalBatches} | 端口 ${batch[0]}-${batch[batch.length-1]} | 进度 ${newProgress}% (${processedCount}/${totalPorts}) ✓`);

            // 短暂延迟，避免过度占用系统资源
            await new Promise(resolve => setTimeout(resolve, 50));

        } catch (error) {
            process.stdout.write(`\r❌ 批次 ${batchNumber}/${totalBatches} | 处理出错: ${error.message}`);
            // 短暂停留显示错误信息
            await new Promise(resolve => setTimeout(resolve, 1000));
            // 继续处理下一批
        }
    }

    // 清除进度条，显示完成信息
    process.stdout.write(`\r✅ 扫描完成！处理了 ${totalPorts} 个端口，共 ${Math.ceil(totalPorts / batchSize)} 个批次\n`);

    return results;
}

async function allTelnetTest(host = 'localhost', maxConcurrency = 100) {
    let ports = []
    for (let i = 1; i <= 65535; i++) {
        ports.push(i)
    }

    console.log(`\n🚀 开始全端口Telnet扫描 ${host}...`);
    console.log(`📋 扫描端口范围: 1-65535 (总共 ${ports.length} 个端口)`);
    console.log(`⚡ 并发限制: ${maxConcurrency} 个连接/批次`);
    console.log('=====================================\n');

    const startTime = Date.now();
    const results = await batchProcessPorts(ports, host, maxConcurrency);

    // 汇总结果
    const summary = {
        open: [],
        closed: [],
        timeout: [],
        error: [],
        host_not_found: []
    };

    results.forEach(result => {
        if (summary[result.status]) {
            summary[result.status].push(result.port);
        }
    });

    // 输出结果
    console.log('\n📊 扫描结果汇总:');
    console.log('=====================================');

    if (summary.open.length > 0) {
        console.log(`🟢 开放端口 (${summary.open.length}): ${summary.open.join(', ')}`);
    }

    if (summary.closed.length > 0) {
        console.log(`🔴 关闭端口 (${summary.closed.length}): 仅显示前20个: ${summary.closed.slice(0,20).join(', ')}${summary.closed.length > 20 ? '...' : ''}`);
    }

    if (summary.timeout.length > 0) {
        console.log(`⏰ 超时端口 (${summary.timeout.length}): 仅显示前20个: ${summary.timeout.slice(0,20).join(', ')}${summary.timeout.length > 20 ? '...' : ''}`);
    }

    if (summary.error.length > 0) {
        console.log(`❌ 错误端口 (${summary.error.length}): 仅显示前20个: ${summary.error.slice(0,20).join(', ')}${summary.error.length > 20 ? '...' : ''}`);
    }

    if (summary.host_not_found.length > 0) {
        console.log(`🔍 主机不存在 (${summary.host_not_found.length}): ${summary.host_not_found.join(', ')}`);
    }

    const endTime = Date.now();
    const duration = ((endTime - startTime) / 1000).toFixed(1);

    console.log(`\n🎯 总计扫描: ${ports.length} 个端口`);
    console.log(`⏱️ 扫描耗时: ${duration} 秒`);
    console.log(`⚡ 扫描完成时间: ${new Date().toLocaleString()}\n`);

    return results;
}

// 命令行接口
async function main() {
    const args = process.argv.slice(2);

    if (args.length === 0) {
        console.log(`测试本机所有端口：1-65535`);
        // 默认测试
        await allTelnetTest("localhost");
        return;
    }
    if (args.length === 1){
        console.log(`测试IP:${args[0]}下的所有端口：1-65535`);
        await allTelnetTest(args[0]);
        return;
    }
    if (args.length === 2){
        console.log(`测试IP:${args[0]}下的端口：${args[1]}`);
        if (args[1] === 'common' || args[1] === 'COMMON' || args[1] === 'c'){
            const defaultPorts = [22,21, 80, 443, 3000, 3306, 5000, 5432, 6379, 8000, 8080,8081,27017];
            await batchTelnetTest(args[0], defaultPorts);
            return;
        }else{
            const ports = args[1].split(',').map(p => parseInt(p.trim())).filter(p => !isNaN(p));
            await batchTelnetTest(args[0], ports);
            if (ports.length === 0) {
                console.error('❌ 无效的端口列表');
                process.exit(1);
            }
            return;
        }
    }
}

// 单个端口测试函数（供外部调用）
async function testSinglePort(host, port) {
    return await telnetPort(host, port);
}

// 如果直接运行此脚本
if (require.main === module) {
    main().catch(error => {
        console.error('❌ 程序执行出错:', error);
        process.exit(1);
    });
}

module.exports = {
    telnetPort,
    batchTelnetTest,
    testSinglePort,
    allTelnetTest
};

