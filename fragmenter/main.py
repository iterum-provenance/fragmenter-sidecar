from pyterum import FragmenterInput, FragmenterOutput

if __name__ == "__main__":
    from_sidecar = FragmenterInput()
    towards_sidecar = FragmenterOutput()

    for file_list in from_sidecar.consumer():
        # If the message is of type KillMessage it returns None, 
        # as to not conflict with the type hint
        if file_list == None:
            towards_sidecar.produce_done()
            break
        for f in file_list:
            towards_sidecar.produce([f])        
    
