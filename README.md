#route
app.Get("/hospitals", getAllHospitals)
app.Get("/hospital/:hospitalId", getHospital)
app.Post("/hospital", newHospital)

app.Get("/groups", getAllGroups)
app.Get("/group/:groupId", getGroup)
app.Post("/group/:hospitalId", addGroup)

app.Get("/sensors", getAllSensors)
app.Get("/sensor/:sensorId", getSensor)
app.Post("/sensor/:groupId", addSensor)

app.Get("/temps", getAllTemps)
app.Get("/temp/:sensorToken", getTemp)
app.Post("/temp", addTemp)
