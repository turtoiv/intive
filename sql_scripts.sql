Create database payments;\
Create table funds(id INT AUTO_INCREMENT PRIMARY KEY, userid INT, funds decimal(8,2));\
Insert into funds values(1,5,5000.4)\
Insert into funds values(2,8,132890.5)\
Insert into funds values(3,11,45000)\
Create table transfers (id INT AUTO_INCREMENT PRIMARY KEY, source INT , destination INT, amount DECIMAL(8,2));\
}