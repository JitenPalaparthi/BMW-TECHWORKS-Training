const largeInt=324324234324234234234234234234234234234234;

let age = 42;

let price = 123.23;

console.log(largeInt);

console.log(typeof largeInt);

console.log(typeof age);

console.log(typeof price);

let a = 10;
let b = "20";

console.log(typeof a);
console.log(typeof b);

console.log(a+b); 

let person=null; // javascript understands it as a object


    let employee={
        "name":"Jiten",
        "age":42,
        "email":"jitenp@outlook.com"
    }

let array=[];

let mymap={};

let notdef;


console.log(typeof person);

console.log(typeof employee);

console.log(typeof array);

console.log(typeof mymap);

console.log(typeof notdef);

let str1 = "123";

str1 = Number(str1)+1
console.log(str1++); // str is converted automatically to a number

console.log(typeof str1);

// arth 
// comparision
// logical


{


var a1 = 10;
var b1 = "10";

if (a1==b1){ // compares only the data but not type
    console.log("both are equal");
}else{
    console.log("both are not equal");
}

if (a1===b1){ // compares  the data and also type
    console.log("both are equal");
}else{
    console.log("both are not equal");
}

if (isNaN(a1)==true){
    console.log("true");
}else{
    console.log("false");
}

str1 = "231a23.23qwe";
let pf = parseFloat(str1);
console.log(typeof pf,pf);

let pi = parseInt(str1);
console.log(typeof pi,pi);

let ch1 = 'A';
let ch2 = "世";
console.log(typeof ch1,typeof ch2);
}

// casting --> Number()
// conversion




