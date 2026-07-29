FROM maven:3.9.9-eclipse-temurin-21

WORKDIR /workspace

COPY pom.xml .
RUN mvn -q -DskipTests dependency:go-offline

COPY src ./src

CMD ["mvn", "clean", "test"]
