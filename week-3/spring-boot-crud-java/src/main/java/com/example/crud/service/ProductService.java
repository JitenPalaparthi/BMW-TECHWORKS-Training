package com.example.crud.service;

import com.example.crud.dto.ProductRequest;
import com.example.crud.dto.ProductResponse;
import com.example.crud.entity.Product;
import com.example.crud.exception.ResourceNotFoundException;
import com.example.crud.repository.ProductRepository;
import java.util.List;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Transactional
public class ProductService {
    private final ProductRepository repository;

    public ProductService(ProductRepository repository) {
        this.repository = repository;
    }

    public ProductResponse create(ProductRequest request) {
        Product product = new Product(
                request.name(), request.description(), request.price(), request.quantity());
        return toResponse(repository.save(product));
    }

    @Transactional(readOnly = true)
    public List<ProductResponse> findAll() {
        return repository.findAll().stream().map(this::toResponse).toList();
    }

    @Transactional(readOnly = true)
    public ProductResponse findById(Long id) {
        return toResponse(findEntity(id));
    }

    public ProductResponse update(Long id, ProductRequest request) {
        Product product = findEntity(id);
        product.setName(request.name());
        product.setDescription(request.description());
        product.setPrice(request.price());
        product.setQuantity(request.quantity());
        return toResponse(repository.save(product));
    }

    public void delete(Long id) {
        Product product = findEntity(id);
        repository.delete(product);
    }

    private Product findEntity(Long id) {
        return repository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Product not found with id: " + id));
    }

    private ProductResponse toResponse(Product product) {
        return new ProductResponse(
                product.getId(), product.getName(), product.getDescription(),
                product.getPrice(), product.getQuantity(),
                product.getCreatedAt(), product.getUpdatedAt());
    }
}
