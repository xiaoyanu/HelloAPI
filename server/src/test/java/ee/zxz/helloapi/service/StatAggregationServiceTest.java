package ee.zxz.helloapi.service;

import ee.zxz.helloapi.mapper.StatAggregationMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InOrder;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.inOrder;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class StatAggregationServiceTest {

    @Mock
    private StatAggregationMapper statAggregationMapper;

    @Test
    void returnsZeroWhenNoNewLogsExist() {
        when(statAggregationMapper.findCheckpoint()).thenReturn(100L);
        when(statAggregationMapper.findBatchUpperLogId(100L, 5000)).thenReturn(100L);
        StatAggregationService service = new StatAggregationService(statAggregationMapper);

        int processed = service.aggregateNextBatch();

        assertEquals(0, processed);
        verify(statAggregationMapper, never()).aggregateDaily(100L, 100L);
        verify(statAggregationMapper, never()).advanceCheckpoint(100L);
    }

    @Test
    void aggregatesBatchBeforeAdvancingCheckpoint() {
        when(statAggregationMapper.findCheckpoint()).thenReturn(100L);
        when(statAggregationMapper.findBatchUpperLogId(100L, 5000)).thenReturn(125L);
        StatAggregationService service = new StatAggregationService(statAggregationMapper);

        int processed = service.aggregateNextBatch();

        assertEquals(25, processed);
        InOrder order = inOrder(statAggregationMapper);
        order.verify(statAggregationMapper).aggregateDaily(100L, 125L);
        order.verify(statAggregationMapper).aggregateTotals(100L, 125L);
        order.verify(statAggregationMapper).advanceCheckpoint(125L);
    }

    @Test
    void doesNotAdvanceCheckpointWhenAggregationFails() {
        when(statAggregationMapper.findCheckpoint()).thenReturn(100L);
        when(statAggregationMapper.findBatchUpperLogId(100L, 5000)).thenReturn(125L);
        org.mockito.Mockito.doThrow(new IllegalStateException("database failure"))
                .when(statAggregationMapper).aggregateDaily(100L, 125L);
        StatAggregationService service = new StatAggregationService(statAggregationMapper);

        assertThrows(IllegalStateException.class, service::aggregateNextBatch);

        verify(statAggregationMapper, never()).advanceCheckpoint(125L);
    }
}
