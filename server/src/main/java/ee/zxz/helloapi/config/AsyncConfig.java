package ee.zxz.helloapi.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

import java.util.concurrent.Executor;
import java.util.concurrent.ThreadPoolExecutor;

@Configuration
@EnableAsync
// 启用每日统计的定时增量聚合任务。
@EnableScheduling
public class AsyncConfig {

    @Bean("logExecutor") // 给线程池起个名字，方便指定
    public Executor logExecutor() {
        ThreadPoolTaskExecutor executor = new ThreadPoolTaskExecutor();
        // 核心线程数：根据你的服务器核心数配置，例如 CPU核数 * 2
        executor.setCorePoolSize(10);
        // 最大线程数
        executor.setMaxPoolSize(20);
        // 队列大小：当线程满了，任务会在队列等待
        executor.setQueueCapacity(500);
        // 线程名称前缀
        executor.setThreadNamePrefix("ApiLog-Async-");
        // 拒绝策略：如果队列也满了，由调用者线程执行（保证日志不丢失，但会轻微阻塞）
        executor.setRejectedExecutionHandler(new ThreadPoolExecutor.CallerRunsPolicy());
        executor.initialize();
        return executor;
    }
}