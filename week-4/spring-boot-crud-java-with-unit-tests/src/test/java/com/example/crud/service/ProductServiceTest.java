package com.example.crud.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.example.crud.dto.ProductRequest;
import com.example.crud.dto.ProductResponse;
import com.example.crud.entity.Product;
import com.example.crud.exception.ResourceNotFoundException;
import com.example.crud.repository.ProductRepository;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.test.util.ReflectionTestUtils;

@ExtendWith(MockitoExtension.class)
class ProductServiceTest {

    @Mock
    private ProductRepository repository;

    private ProductService service;

    @BeforeEach
    void setUp() {
        service = new ProductService(repository);
    }

    @Test
    void createShouldSaveAndReturnProduct() {
        ProductRequest request = request("Laptop", "Developer laptop", "125000.00", 4);
        when(repository.save(any(Product.class))).thenAnswer(invocation -> {
            Product product = invocation.getArgument(0);
            setPersistenceFields(product, 1L);
            return product;
        });

        ProductResponse response = service.create(request);

        ArgumentCaptor<Product> captor = ArgumentCaptor.forClass(Product.class);
        verify(repository).save(captor.capture());
        Product saved = captor.getValue();

        assertThat(saved.getName()).isEqualTo("Laptop");
        assertThat(saved.getDescription()).isEqualTo("Developer laptop");
        assertThat(saved.getPrice()).isEqualByComparingTo("125000.00");
        assertThat(saved.getQuantity()).isEqualTo(4);

        assertThat(response.id()).isEqualTo(1L);
        assertThat(response.name()).isEqualTo("Laptop");
        assertThat(response.price()).isEqualByComparingTo("125000.00");
    }

    @Test
    void findAllShouldMapEntitiesToResponses() {
        Product first = product(1L, "Keyboard", "Mechanical", "4500.00", 10);
        Product second = product(2L, "Mouse", "Wireless", "2200.00", 7);
        when(repository.findAll()).thenReturn(List.of(first, second));

        List<ProductResponse> responses = service.findAll();

        assertThat(responses)
                .extracting(ProductResponse::name)
                .containsExactly("Keyboard", "Mouse");
        assertThat(responses)
                .extracting(ProductResponse::id)
                .containsExactly(1L, 2L);
        verify(repository).findAll();
    }

    @Test
    void findByIdShouldReturnProductWhenPresent() {
        Product product = product(10L, "Monitor", "4K monitor", "32000.00", 3);
        when(repository.findById(10L)).thenReturn(Optional.of(product));

        ProductResponse response = service.findById(10L);

        assertThat(response.id()).isEqualTo(10L);
        assertThat(response.name()).isEqualTo("Monitor");
        assertThat(response.quantity()).isEqualTo(3);
    }

    @Test
    void findByIdShouldThrowWhenProductDoesNotExist() {
        when(repository.findById(99L)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.findById(99L))
                .isInstanceOf(ResourceNotFoundException.class)
                .hasMessage("Product not found with id: 99");
    }

    @Test
    void updateShouldModifyAndSaveExistingProduct() {
        Product existing = product(5L, "Old name", "Old description", "10.00", 1);
        ProductRequest request = request("New name", "New description", "25.50", 8);
        when(repository.findById(5L)).thenReturn(Optional.of(existing));
        when(repository.save(existing)).thenReturn(existing);

        ProductResponse response = service.update(5L, request);

        assertThat(existing.getName()).isEqualTo("New name");
        assertThat(existing.getDescription()).isEqualTo("New description");
        assertThat(existing.getPrice()).isEqualByComparingTo("25.50");
        assertThat(existing.getQuantity()).isEqualTo(8);
        assertThat(response.name()).isEqualTo("New name");
        verify(repository).save(existing);
    }

    @Test
    void updateShouldNotSaveWhenProductDoesNotExist() {
        ProductRequest request = request("New name", "Description", "25.50", 8);
        when(repository.findById(404L)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.update(404L, request))
                .isInstanceOf(ResourceNotFoundException.class)
                .hasMessage("Product not found with id: 404");

        verify(repository, never()).save(any(Product.class));
    }

    @Test
    void deleteShouldDeleteExistingProduct() {
        Product product = product(7L, "SSD", "1 TB", "7000.00", 6);
        when(repository.findById(7L)).thenReturn(Optional.of(product));

        service.delete(7L);

        verify(repository).delete(product);
    }

    @Test
    void deleteShouldNotCallRepositoryDeleteWhenProductDoesNotExist() {
        when(repository.findById(8L)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> service.delete(8L))
                .isInstanceOf(ResourceNotFoundException.class)
                .hasMessage("Product not found with id: 8");

        verify(repository, never()).delete(any(Product.class));
    }

    private static ProductRequest request(
            String name, String description, String price, int quantity) {
        return new ProductRequest(name, description, new BigDecimal(price), quantity);
    }

    private static Product product(
            Long id, String name, String description, String price, int quantity) {
        Product product = new Product(name, description, new BigDecimal(price), quantity);
        setPersistenceFields(product, id);
        return product;
    }

    private static void setPersistenceFields(Product product, Long id) {
        LocalDateTime timestamp = LocalDateTime.of(2026, 8, 4, 10, 0);
        ReflectionTestUtils.setField(product, "id", id);
        ReflectionTestUtils.setField(product, "createdAt", timestamp);
        ReflectionTestUtils.setField(product, "updatedAt", timestamp);
    }
}
