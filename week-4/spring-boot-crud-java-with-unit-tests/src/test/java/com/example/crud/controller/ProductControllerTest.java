package com.example.crud.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.example.crud.dto.ProductRequest;
import com.example.crud.dto.ProductResponse;
import com.example.crud.service.ProductService;
import java.math.BigDecimal;
import java.net.URI;
import java.time.LocalDateTime;
import java.util.List;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

@ExtendWith(MockitoExtension.class)
class ProductControllerTest {

    @Mock
    private ProductService service;

    private ProductController controller;

    @BeforeEach
    void setUp() {
        controller = new ProductController(service);
    }

    @AfterEach
    void tearDown() {
        RequestContextHolder.resetRequestAttributes();
    }

    @Test
    void createShouldReturn201WithLocationAndBody() {
        ProductRequest request = request();
        ProductResponse created = response(11L, "Laptop");
        when(service.create(request)).thenReturn(created);

        MockHttpServletRequest servletRequest = new MockHttpServletRequest("POST", "/api/products");
        servletRequest.setScheme("http");
        servletRequest.setServerName("localhost");
        servletRequest.setServerPort(8080);
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(servletRequest));

        ResponseEntity<ProductResponse> result = controller.create(request);

        assertThat(result.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        assertThat(result.getHeaders().getLocation())
                .isEqualTo(URI.create("http://localhost:8080/api/products/11"));
        assertThat(result.getBody()).isEqualTo(created);
        verify(service).create(request);
    }

    @Test
    void findAllShouldReturnServiceResults() {
        List<ProductResponse> expected = List.of(response(1L, "Keyboard"), response(2L, "Mouse"));
        when(service.findAll()).thenReturn(expected);

        List<ProductResponse> result = controller.findAll();

        assertThat(result).isSameAs(expected);
        verify(service).findAll();
    }

    @Test
    void findByIdShouldDelegateToService() {
        ProductResponse expected = response(3L, "Monitor");
        when(service.findById(3L)).thenReturn(expected);

        ProductResponse result = controller.findById(3L);

        assertThat(result).isEqualTo(expected);
        verify(service).findById(3L);
    }

    @Test
    void updateShouldDelegateToService() {
        ProductRequest request = request();
        ProductResponse expected = response(4L, "Laptop");
        when(service.update(4L, request)).thenReturn(expected);

        ProductResponse result = controller.update(4L, request);

        assertThat(result).isEqualTo(expected);
        verify(service).update(4L, request);
    }

    @Test
    void deleteShouldReturn204AndDelegateToService() {
        ResponseEntity<Void> result = controller.delete(5L);

        assertThat(result.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
        assertThat(result.getBody()).isNull();
        verify(service).delete(5L);
    }

    private static ProductRequest request() {
        return new ProductRequest(
                "Laptop", "Developer laptop", new BigDecimal("125000.00"), 4);
    }

    private static ProductResponse response(Long id, String name) {
        LocalDateTime timestamp = LocalDateTime.of(2026, 8, 4, 10, 0);
        return new ProductResponse(
                id,
                name,
                "Description",
                new BigDecimal("100.00"),
                2,
                timestamp,
                timestamp);
    }
}
