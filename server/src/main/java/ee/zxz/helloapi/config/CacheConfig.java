package ee.zxz.helloapi.config;

import com.github.benmanes.caffeine.cache.Caffeine;
import org.springframework.cache.CacheManager;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.cache.caffeine.CaffeineCacheManager;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.concurrent.TimeUnit;

/** Dashboard 本地缓存配置。 */
@Configuration
@EnableCaching
public class CacheConfig {
    /**
     * 创建统计缓存管理器。
     *
     * <p>最多保存 10 个键，每个 Dashboard 结果写入 30 秒后失效。</p>
     *
     * @return 使用 Caffeine 实现的 Spring CacheManager
     */
    @Bean
    public CacheManager cacheManager() {
        CaffeineCacheManager cacheManager = new CaffeineCacheManager("statDashboard", "apiAllCount");
        cacheManager.setCaffeine(Caffeine.newBuilder()
                .maximumSize(10)
                .expireAfterWrite(30, TimeUnit.SECONDS));
        return cacheManager;
    }
}
