
let numbers = [10,20,30,40,50];

let array2d=[[10,20],[30,40],[50,60]];

console.log(numbers);

console.log(array2d);

let fruits = ["apple","banana","mango","kiwi"];

numbers.push(60);

console.log(numbers);

array2d.push([70,80]);

console.log(array2d);

let sum =0;
for(let i=0;i<numbers.length;i++){
sum+=numbers[i];
}
console.log("Sum:",sum);

numbers.forEach(function(value){
sum+=value;
console.log(value);
});

console.log("Sum:",sum);


let doubles= numbers.map(n=>n*2);

console.log(doubles);


sum=0;

array2d.forEach(function(arr){
    arr.forEach(function(value){
    sum +=value
    });
})



console.log("using foreach loop on a collection",sum);


sum=0;
for(let i=0;i<array2d.length;i++){
    for(let j=0;j<array2d[i].length;j++){
        sum+=array2d[i][j];
    }
}
console.log("using normal nested loop",sum);


let arr1=[5,2,9,1];


arr1.sort((a,b)=>a-b);

console.log(arr1);

console.log(fruits.reverse());